package kvdb2ethdb

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/ethdb"
)

type Adapter struct {
	kvdb.Store
}

var _ ethdb.KeyValueStore = (*Adapter)(nil)

func Wrap(v kvdb.Store) *Adapter {
	return &Adapter{v}
}

// batch is a write-only memory batch that commits changes to its host
// database when Write is called. A batch cannot be used concurrently.
type batch struct {
	kvdb.Batch
}

// Replay replays the batch contents.
func (b *batch) Replay(w ethdb.KeyValueWriter) error {
	return b.Batch.Replay(w)
}

// NewBatch creates a write-only key-value store that buffers changes to its host
// database until a final write is called.
func (db *Adapter) NewBatch() ethdb.Batch {
	return &batch{db.Store.NewBatch()}
}

// NewIterator creates a binary-alphabetical iterator over the entire keyspace
// contained within the memory database.
func (db *Adapter) NewIterator() ethdb.Iterator {
	return db.NewIterator()
}

// NewIteratorWithStart creates a binary-alphabetical iterator over a subset of
// database content starting at a particular initial key (or after, if it does
// not exist).
func (db *Adapter) NewIteratorWithStart(start []byte) ethdb.Iterator {
	return db.NewIteratorWithStart(start)
}

// NewIteratorWithPrefix creates a binary-alphabetical iterator over a subset
// of database content with a particular key prefix.
func (db *Adapter) NewIteratorWithPrefix(prefix []byte) ethdb.Iterator {
	return db.NewIteratorWithPrefix(prefix)
}
