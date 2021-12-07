package snap2kvdb

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
)

type Adapter struct {
	kvdb.Snapshot
}

var _ kvdb.Store = (*Adapter)(nil)

func Wrap(v kvdb.Snapshot) *Adapter {
	return &Adapter{v}
}

func (db *Adapter) Put(key []byte, value []byte) error {
	panic("called Put on snapshot")
	return nil
}

func (db *Adapter) Delete(key []byte) error {
	panic("called Delete on snapshot")
	return nil
}

func (db *Adapter) GetSnapshot() (kvdb.Snapshot, error) {
	return db.Snapshot, nil
}

func (db *Adapter) NewBatch() kvdb.Batch {
	panic("called NewBatch on snapshot")
	return nil
}

func (db *Adapter) Compact(start []byte, limit []byte) error {
	return nil
}

func (db *Adapter) Close() error {
	return nil
}

func (db *Adapter) Stat(property string) (string, error) {
	return "", nil
}
