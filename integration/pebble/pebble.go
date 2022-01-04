package pebble

import (
	"fmt"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/cockroachdb/pebble"
	"sync"
)

// Database is a persistent key-value store. Apart from basic data storage
// functionality it also supports batch writes and iterating over the keyspace in
// binary-alphabetical order.
type Database struct {
	fn string     // filename for reporting
	db *pebble.DB // LevelDB instance

	quitLock sync.Mutex // Mutex protecting the quit channel access

	onClose func() error
	onDrop  func()
}

// New returns a wrapped LevelDB object. The namespace is the prefix that the
// metrics reporting should use for surfacing internal stats.
func New(path string, close func() error, drop func()) (*Database, error) {
	db, err := pebble.Open(path, &pebble.Options{
		BytesPerSync:                512 << 10, // SSTable syncs (512 KB)
		Cache:                       pebble.NewCache(8 << 20), // 8 MB
		L0CompactionThreshold:       4, // default: 4
		L0StopWritesThreshold:       12, // default: 12
		LBaseMaxBytes:               64 << 20, // default: 64 MB
		MaxManifestFileSize:         128 << 20, // default: 128 MB
		MaxOpenFiles:                1000,
		MemTableSize:                4 << 20, // default: 4 MB
		MemTableStopWritesThreshold: 2, // writes are stopped when sum of the queued memtable sizes exceeds
		MaxConcurrentCompactions:    1,
		NumPrevManifest:             1, // keep one old manifest
		WALBytesPerSync:             0, // default 0 (matches RocksDB)
	})

	if err != nil {
		return nil, err
	}
	// Assemble the wrapper with all the registered metrics
	ldb := Database{
		fn: path,
		db: db,
		onClose: close,
		onDrop: drop,
	}
	return &ldb, nil
}

// Close stops the metrics collection, flushes any pending data to disk and closes
// all io accesses to the underlying key-value store.
func (db *Database) Close() error {
	db.quitLock.Lock()
	defer db.quitLock.Unlock()

	if db.db == nil {
		panic("already closed")
	}

	ldb := db.db
	db.db = nil

	if db.onClose != nil {
		if err := db.onClose(); err != nil {
			return err
		}
		db.onClose = nil
	}
	if err := ldb.Close(); err != nil {
		return err
	}
	return nil
}

// Drop whole database.
func (db *Database) Drop() {
	if db.db != nil {
		panic("Close database first!")
	}
	if db.onDrop != nil {
		db.onDrop()
	}
}

// Has retrieves if a key is present in the key-value store.
func (db *Database) Has(key []byte) (bool, error) {
	_, closer, err := db.db.Get(key)
	if err == pebble.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = closer.Close()
	return true, err
}

// Get retrieves the given key if it's present in the key-value store.
func (db *Database) Get(key []byte) ([]byte, error) {
	value, closer, err := db.db.Get(key)
	if err == pebble.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	clonedValue := append([]byte{}, value...)
	err = closer.Close()
	return clonedValue, err
}

// Put inserts the given value into the key-value store.
func (db *Database) Put(key []byte, value []byte) error {
	return db.db.Set(key, value, pebble.NoSync)
}

// Delete removes the key from the key-value store.
func (db *Database) Delete(key []byte) error {
	return db.db.Delete(key, pebble.NoSync)
}

// NewBatch creates a write-only key-value store that buffers changes to its host
// database until a final write is called.
func (db *Database) NewBatch() kvdb.Batch {
	return &batch{
		db: db.db,
		b:  db.db.NewBatch(),
	}
}

// NewIterator creates a binary-alphabetical iterator over a subset
// of database content with a particular key prefix, starting at a particular
// initial key (or after, if it does not exist).
func (db *Database) NewIterator(prefix []byte, start []byte) kvdb.Iterator {
	x := iterator{db.db.NewIter(bytesPrefixRange(prefix, start)), false, false}
	return &x
}

type iterator struct {
	*pebble.Iterator
	isStarted bool
	isClosed bool
}

func (it *iterator) Next() bool {
	if it.isStarted {
		return it.Iterator.Next()
	} else {
		// pebble needs First() instead of the first Next()
		it.isStarted = true
		return it.Iterator.First()
	}
}

func (it *iterator) Release() {
	if it.isClosed {
		return
	}
	_ = it.Iterator.Close() // must not be called multiple times
	it.isClosed = true
}

// bytesPrefixRange returns key range that satisfy
// - the given prefix, and
// - the given seek position
func bytesPrefixRange(prefix, start []byte) *pebble.IterOptions {
	if prefix == nil && start == nil {
		return nil
	}
	var r pebble.IterOptions
	if prefix != nil {
		r = bytesPrefix(prefix)
	} else {
		r.LowerBound = []byte{}
	}
	r.LowerBound = append(r.LowerBound, start...)
	return &r
}

// bytesPrefix is copied from leveldb util
func bytesPrefix(prefix []byte) pebble.IterOptions {
	var limit []byte
	for i := len(prefix) - 1; i >= 0; i-- {
		c := prefix[i]
		if c < 0xff {
			limit = make([]byte, i+1)
			copy(limit, prefix)
			limit[i] = c + 1
			break
		}
	}
	return pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: limit,
	}
}

// Stat returns a particular internal stat of the database.
func (db *Database) Stat(property string) (string, error) {
	if property == "leveldb.iostats" {
		return fmt.Sprintf("Read(MB):%.5f Write(MB):%.5f",
			float64(db.db.Metrics().Total().BytesRead)/1048576.0, // 1024*1024
			float64(db.db.Metrics().Total().BytesIn)/1048576.0), nil
	}
	if property == "leveldb.metrics" {
		return db.db.Metrics().String(), nil
	}
	return "", fmt.Errorf("pebble stat property %s does not exists", property)
}

// Compact flattens the underlying data store for the given key range. In essence,
// deleted and overwritten versions are discarded, and the data is rearranged to
// reduce the cost of operations needed to access them.
//
// A nil start is treated as a key before all keys in the data store; a nil limit
// is treated as a key after all keys in the data store. If both is nil then it
// will compact entire data store.
func (db *Database) Compact(start []byte, limit []byte) error {
	if limit == nil {
		limit = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	}
	return db.db.Compact(start, limit, true)
}

// Path returns the path to the database directory.
func (db *Database) Path() string {
	return db.fn
}

// GetSnapshot returns the latest snapshot of the underlying DB. A snapshot
// is a frozen snapshot of a DB state at a particular point in time. The
// content of snapshot are guaranteed to be consistent.
//
// The snapshot must be released after use, by calling Release method.
func (db *Database) GetSnapshot() (kvdb.Snapshot, error) {
	return &snapshot{
		db:   db.db,
		snap: db.db.NewSnapshot(),
	}, nil
}

type snapshot struct {
	db   *pebble.DB
	snap *pebble.Snapshot
}

// Has retrieves if a key is present in the key-value store.
func (s *snapshot) Has(key []byte) (bool, error) {
	_, closer, err := s.snap.Get(key)
	if err == pebble.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = closer.Close()
	return true, err
}

// Get retrieves the given key if it's present in the key-value store.
func (s *snapshot) Get(key []byte) ([]byte, error) {
	value, closer, err := s.snap.Get(key)
	if err == pebble.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	clonedValue := append([]byte{}, value...)
	err = closer.Close()
	return clonedValue, err
}

func (s *snapshot) NewIterator(prefix []byte, start []byte) kvdb.Iterator {
	x := iterator{s.snap.NewIter(bytesPrefixRange(prefix, start)), false, false}
	return &x
}

func (s *snapshot) Release() {
	_ = s.snap.Close()
}

// batch is a write-only leveldb batch that commits changes to its host database
// when Write is called. A batch cannot be used concurrently.
type batch struct {
	db   *pebble.DB
	b    *pebble.Batch
	size int
}

// Put inserts the given value into the batch for later committing.
func (b *batch) Put(key, value []byte) error {
	err := b.b.Set(key, value, pebble.NoSync)
	b.size += len(value)
	return err
}

// Delete inserts the key removal into the batch for later committing.
func (b *batch) Delete(key []byte) error {
	err := b.b.Delete(key, pebble.NoSync)
	b.size++
	return err
}

// ValueSize retrieves the amount of data queued up for writing.
func (b *batch) ValueSize() int {
	return b.size
}

// Write flushes any accumulated data to disk.
func (b *batch) Write() error {
	return b.db.Apply(b.b, pebble.NoSync)
}

// Reset resets the batch for reuse.
func (b *batch) Reset() {
	b.b.Reset()
	b.size = 0
}

// Replay replays the batch contents.
func (b *batch) Replay(w kvdb.Writer) (err error) {
	for iter := b.b.Reader(); len(iter) > 0; {
		kind, key, value, ok := iter.Next()
		if !ok {
			break
		}
		switch kind {
		case pebble.InternalKeyKindSet:
			err = w.Put(key, value)
		case pebble.InternalKeyKindDelete:
			err = w.Delete(key)
		}
		if err != nil {
			break
		}
	}
	return
}
