package evmstore

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
)

func cachedStore() *Store {
	mems := memorydb.NewProducer("", withDelay)
	dbs, err := mems.OpenDB("test")
	if err != nil {
		panic(err)
	}
	cfg := LiteStoreConfig()
	return NewStore(dbs, cfg)
}

func nonCachedStore() *Store {
	mems := memorydb.NewProducer("", withDelay)
	dbs, err := mems.OpenDB("test")
	if err != nil {
		panic(err)
	}
	cfg := StoreConfig{}

	return NewStore(dbs, cfg)
}

func withDelay(db kvdb.DropableStore) kvdb.DropableStore {
	mem, ok := db.(*memorydb.Database)
	if ok {
		mem.SetDelay(time.Millisecond)

	}

	return db
}
