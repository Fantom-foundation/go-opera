package evmstore

import (
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
)

func cachedStore() *Store {
	cfg := LiteStoreConfig()

	return NewStore(memorydb.NewProducer(""), cfg)
}

func nonCachedStore() *Store {
	cfg := StoreConfig{}

	return NewStore(memorydb.NewProducer(""), cfg)
}

func TestStoreSetTx(t *testing.T) {
	store := cachedStore()

	tx := types.NewTx(&types.LegacyTx{Data: []byte("test")})

	store.SetTx(tx.Hash(), tx)

	txFromStore := store.GetTx(tx.Hash())
	assert.Equal(t, tx.Data(), txFromStore.Data())
}
