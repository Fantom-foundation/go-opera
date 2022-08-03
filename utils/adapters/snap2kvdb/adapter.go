package snap2kvdb

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/devnulldb"
	"github.com/ethereum/go-ethereum/log"
)

type Adapter struct {
	kvdb.Snapshot
}

var _ kvdb.Store = (*Adapter)(nil)

func Wrap(v kvdb.Snapshot) *Adapter {
	return &Adapter{v}
}

func (db *Adapter) Put(key []byte, value []byte) error {
	log.Warn("called Put on snapshot")
	return nil
}

func (db *Adapter) Delete(key []byte) error {
	log.Warn("called Delete on snapshot")
	return nil
}

func (db *Adapter) GetSnapshot() (kvdb.Snapshot, error) {
	return db.Snapshot, nil
}

func (db *Adapter) NewBatch() kvdb.Batch {
	log.Warn("called NewBatch on snapshot")
	return devnulldb.New().NewBatch()
}

func (db *Adapter) Compact(start []byte, limit []byte) error {
	return nil
}

func (db *Adapter) Close() error {
	return nil
}

func (db *Adapter) Drop() {}

func (db *Adapter) Stat(property string) (string, error) {
	return "", nil
}
