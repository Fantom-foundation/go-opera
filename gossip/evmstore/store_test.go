package evmstore

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
)

func cachedStore() *Store {
	mems := memorydb.NewProducer("", withDelay)
	cfg := LiteStoreConfig()

	return NewStore(mems.OpenDb("test"), cfg)
}

func nonCachedStore() *Store {
	mems := memorydb.NewProducer("", withDelay)
	cfg := StoreConfig{}

	return NewStore(mems.OpenDb("test"), cfg)
}

func withDelay(db kvdb.DropableStore) kvdb.DropableStore {
	mem, ok := db.(*memorydb.Database)
	if ok {
		mem.SetDelay(time.Millisecond)

	}

	return db
}
