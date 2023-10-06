package evmstore

import (
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

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
	require.Equal(t, tx.Hash(), txFromStore.Hash())
	require.Equal(t, tx.Data(), txFromStore.Data())
	require.Equal(t, tx.Nonce(), txFromStore.Nonce())
	require.Equal(t, tx.Size(), txFromStore.Size())
	require.Equal(t, tx.Value(), txFromStore.Value())
	require.Equal(t, tx.To(), txFromStore.To())
	require.Equal(t, tx.Gas(), txFromStore.Gas())
	require.Equal(t, tx.GasPrice(), txFromStore.GasPrice())
}
