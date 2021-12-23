package evmstore

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
)

func cachedStore() *Store {
	return NewStore(memorydb.New(), LiteStoreConfig())
}

func nonCachedStore() *Store {
	return NewStore(memorydb.New(), StoreConfig{})
}
