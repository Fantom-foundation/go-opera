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

func (db *Adapter) Drop() {
	panic("called Drop on ethdb")
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

func (db *Adapter) GetSnapshot() (kvdb.Snapshot, error) {
	panic("called GetSnapshot on ethdb")
	return nil, nil
}

// NewIterator creates a binary-alphabetical iterator over a subset
// of database content with a particular key prefix, starting at a particular
// initial key (or after, if it does not exist).
func (db *Adapter) NewIterator(prefix []byte, start []byte) kvdb.Iterator {
	return db.KeyValueStore.NewIterator(prefix, start)
}
