package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
)

func cachedStore() *Store {
	mems := memorydb.NewProdicer("", withDelay)
	dbs := flushable.NewSyncedPool(mems)
	cfg := LiteStoreConfig()

	return NewStore(dbs, cfg)
}

func nonCachedStore() *Store {
	mems := memorydb.NewProdicer("", withDelay)
	dbs := flushable.NewSyncedPool(mems)
	cfg := StoreConfig{}

	return NewStore(dbs, cfg)
}

func withDelay(db kvdb.KeyValueStore) kvdb.KeyValueStore {
	// TODO: uncomment
	/*
		mem, ok := db.(*memorydb.Database)
		if ok {
			 mem.SetDelay(time.Millisecond)

		}
	*/

	return db
}
