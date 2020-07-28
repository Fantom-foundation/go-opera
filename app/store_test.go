package app

import (
	"time"

	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
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

func withDelay(db kvdb.KeyValueStore) kvdb.KeyValueStore {
	mem, ok := db.(*memorydb.Database)
	if ok {
		mem.SetDelay(time.Millisecond)

	}

	return db
}
