package app

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

func TestStoreGetReceipts(t *testing.T) {
	logger.SetTestMode(t)

	block, expect := fakeReceipts()
	store := cachedStore()
	store.SetReceipts(block, *expect)

	got := store.GetReceipts(block)
	assert.EqualValues(t, expect, &got)
}

func BenchmarkStoreGetReceipts(b *testing.B) {
	logger.SetTestMode(b)

	b.Run("cache on", func(b *testing.B) {
		benchStoreGetReceipts(b, cachedStore())
	})
	b.Run("cache off", func(b *testing.B) {
		benchStoreGetReceipts(b, nonCachedStore())
	})
}

func benchStoreGetReceipts(b *testing.B, store *Store) {
	block, receipt := fakeReceipts()

	store.SetReceipts(block, *receipt)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if store.GetReceipts(block) == nil {
			b.Fatal("invalid result")
		}
	}
}

func BenchmarkStoreSetReceipts(b *testing.B) {
	logger.SetTestMode(b)

	b.Run("cache on", func(b *testing.B) {
		benchStoreSetReceipts(b, cachedStore())
	})
	b.Run("cache off", func(b *testing.B) {
		benchStoreSetReceipts(b, nonCachedStore())
	})
}

func benchStoreSetReceipts(b *testing.B, store *Store) {
	block, receipt := fakeReceipts()

	for i := 0; i < b.N; i++ {
		store.SetReceipts(block, *receipt)
	}
}

func fakeReceipts() (idx.Block, *types.Receipts) {
	return idx.Block(1),
		&types.Receipts{
			&types.Receipt{
				PostState:         nil,
				Status:            0,
				CumulativeGasUsed: 0,
				Bloom:             types.Bloom{},
				Logs:              nil,
				TxHash:            common.Hash{},
				ContractAddress:   common.Address{},
				GasUsed:           0,
				BlockHash:         common.Hash{},
				BlockNumber:       nil,
				TransactionIndex:  0,
			},
		}
}
