package gossip

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestStoreGetReceipts(t *testing.T) {
	store := _lruStore()

	expect := _createTestReceipts()

	store.SetReceipts(1, *expect)
	got := store.GetReceipts(1)

	assert.EqualValues(t, expect, &got)
}

func BenchmarkReadReceipts(b *testing.B) {
	b.Run("LRU on", func(b *testing.B) {
		benchReadReceipts(b, _lruStore())
	})
	b.Run("LRU off", func(b *testing.B) {
		benchReadReceipts(b, _simpleStore())
	})
}

func benchReadReceipts(b *testing.B, store *Store) {
	expect := _createTestReceipts()

	if store.cache.Receipts != nil {
		store.cache.Receipts.Purge()
	}

	store.SetReceipts(1, *expect)

	for i := 0; i < b.N; i++ {
		_ = store.GetReceipts(1)
	}
}

func BenchmarkWriteReceipts(b *testing.B) {
	b.Run("LRU on", func(b *testing.B) {
		benchWriteReceipts(b, _lruStore())
	})
	b.Run("LRU off", func(b *testing.B) {
		benchWriteReceipts(b, _simpleStore())
	})
}

func benchWriteReceipts(b *testing.B, store *Store) {
	expect := _createTestReceipts()

	if store.cache.Receipts != nil {
		store.cache.Receipts.Purge()
	}

	for i := 0; i < b.N; i++ {
		store.SetReceipts(1, *expect)
	}
}

func _createTestReceipts() *types.Receipts {
	d := &types.Receipts{
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

	return d
}
