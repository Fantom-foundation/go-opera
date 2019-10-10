package topicsdb

import (
	"github.com/Fantom-foundation/go-lachesis/kvdb"
)

type TopicsDb struct {
	db kvdb.KeyValueStore
}

func New(db kvdb.KeyValueStore) *TopicsDb {
	return &TopicsDb{
		db: db,
	}
}
