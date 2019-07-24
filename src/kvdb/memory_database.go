package kvdb

import (
	"bytes"
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// MemDatabase is a kvbd.Database wrapper of map[string][]byte
// Do not use for any production it does not get persisted
type MemDatabase struct {
	db     map[string][]byte
	prefix []byte
	lock   *sync.RWMutex
}

// NewMemDatabase wraps map[string][]byte
func NewMemDatabase() *MemDatabase {
	return &MemDatabase{
		db:   make(map[string][]byte),
		lock: new(sync.RWMutex),
	}
}

/*
 * Database interface implementation
 */

// NewTable returns a Database object that prefixes all keys with a given prefix.
func (w *MemDatabase) NewTable(prefix []byte) Database {
	return &MemDatabase{
		db:     w.db,
		prefix: prefix,
		lock:   w.lock,
	}
}

// Put puts key-value pair into db.
func (w *MemDatabase) Put(key []byte, value []byte) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	key = append(w.prefix, key...)

	w.db[string(key)] = common.CopyBytes(value)
	return nil
}

// Has checks if key is in the db.
func (w *MemDatabase) Has(key []byte) (bool, error) {
	w.lock.RLock()
	defer w.lock.RUnlock()

	key = append(w.prefix, key...)

	_, ok := w.db[string(key)]
	return ok, nil
}

// Get returns key-value pair by key.
func (w *MemDatabase) Get(key []byte) ([]byte, error) {
	w.lock.RLock()
	defer w.lock.RUnlock()

	key = append(w.prefix, key...)

	if entry, ok := w.db[string(key)]; ok {
		return common.CopyBytes(entry), nil
	}
	return nil, nil
}

// ForEach scans key-value pair by key prefix.
func (w *MemDatabase) ForEach(prefix []byte, do func(key, val []byte)) error {
	w.lock.RLock()
	defer w.lock.RUnlock()

	for k, val := range w.db {
		key := []byte(k)
		if bytes.HasPrefix(key, prefix) {
			do(key, val)
		}
	}

	return nil
}

// Delete removes key-value pair by key.
func (w *MemDatabase) Delete(key []byte) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	key = append(w.prefix, key...)

	delete(w.db, string(key))
	return nil
}

// Close leaves underlying database.
func (w *MemDatabase) Close() {
	w.db = nil
}

// NewBatch creates new batch.
func (w *MemDatabase) NewBatch() Batch {
	return &memBatch{db: w}
}

/*
 * Batch
 */

type kv struct {
	k, v []byte
	del  bool
}

// memBatch is a batch structure.
type memBatch struct {
	db     *MemDatabase
	writes []kv
	size   int
}

// Put puts key-value pair into batch.
func (b *memBatch) Put(key, value []byte) error {
	key = append(b.db.prefix, key...)

	b.writes = append(b.writes, kv{common.CopyBytes(key), common.CopyBytes(value), false})
	b.size += len(value)
	return nil
}

// Delete removes key-value pair from batch by key.
func (b *memBatch) Delete(key []byte) error {
	key = append(b.db.prefix, key...)

	b.writes = append(b.writes, kv{common.CopyBytes(key), nil, true})
	b.size++
	return nil
}

// Write writes batch into db.
func (b *memBatch) Write() error {
	b.db.lock.Lock()
	defer b.db.lock.Unlock()

	for _, kv := range b.writes {
		if kv.del {
			delete(b.db.db, string(kv.k))
			continue
		}
		b.db.db[string(kv.k)] = kv.v
	}
	return nil
}

// ValueSize returns values sizes sum.
func (b *memBatch) ValueSize() int {
	return b.size
}

// Reset cleans whole batch.
func (b *memBatch) Reset() {
	b.writes = b.writes[:0]
	b.size = 0
}

/*
 * for debug:
 */

// Keys returns slice of keys in the db.
func (w *MemDatabase) Keys() [][]byte {
	w.lock.RLock()
	defer w.lock.RUnlock()

	keys := [][]byte{}
	for key := range w.db {
		keys = append(keys, []byte(key))
	}
	return keys
}

// Len returns count of key-values pairs in the db.
func (w *MemDatabase) Len() int {
	w.lock.RLock()
	defer w.lock.RUnlock()

	return len(w.db)
}
