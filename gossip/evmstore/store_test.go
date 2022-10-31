package evmstore

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
)

func cachedStore() *Store {
	cfg := LiteStoreConfig()

	return NewStore(memorydb.New(), cfg, nil, nil)
}

func nonCachedStore() *Store {
	cfg := StoreConfig{}

	return NewStore(memorydb.New(), cfg, nil, nil)
}
