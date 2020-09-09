package ethdb2kvdb

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/ethdb"
)

type Adapter struct {
	ethdb.KeyValueStore
}

var _ kvdb.Store = (*Adapter)(nil)

func Wrap(v ethdb.KeyValueStore) *Adapter {
	return &Adapter{v}
}

// batch is a write-only memory batch that commits changes to its host
// database when Write is called. A batch cannot be used concurrently.
type batch struct {
	ethdb.Batch
}

// Replay replays the batch contents.
func (b *batch) Replay(w kvdb.Writer) error {
	return b.Batch.Replay(w)
}

// NewBatch creates a write-only key-value store that buffers changes to its host
// database until a final write is called.
func (db *Adapter) NewBatch() kvdb.Batch {
	return &batch{db.KeyValueStore.NewBatch()}
}

// NewIterator creates a binary-alphabetical iterator over the entire keyspace
// contained within the memory database.
func (db *Adapter) NewIterator() kvdb.Iterator {
	return db.NewIterator()
}

// NewIteratorWithStart creates a binary-alphabetical iterator over a subset of
// database content starting at a particular initial key (or after, if it does
// not exist).
func (db *Adapter) NewIteratorWithStart(start []byte) kvdb.Iterator {
	return db.NewIteratorWithStart(start)
}

// NewIteratorWithPrefix creates a binary-alphabetical iterator over a subset
// of database content with a particular key prefix.
func (db *Adapter) NewIteratorWithPrefix(prefix []byte) kvdb.Iterator {
	return db.NewIteratorWithPrefix(prefix)
}
