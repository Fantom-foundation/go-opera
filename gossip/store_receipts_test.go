package gossip

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestStoreGetReceipts(t *testing.T) {
	store := lruStore

	expect := _createTestReceipts()

	store.SetReceipts(1, *expect)
	got := store.GetReceipts(1)

	assert.EqualValues(t, expect, &got)
}

func BenchmarkReadReceipts(b *testing.B) {
	testStore = lruStore
	b.Run("LRUon", benchReadReceipts)

	testStore = simpleStore
	b.Run("LRUoff", benchReadReceipts)
}

func benchReadReceipts(b *testing.B) {
	expect := _createTestReceipts()

	if testStore.cache.Receipts != nil {
		testStore.cache.Receipts.Purge()
	}

	testStore.SetReceipts(1, *expect)

	for i := 0; i < b.N; i++ {
		_ = testStore.GetReceipts(1)
	}
}

func BenchmarkWriteReceipts(b *testing.B) {
	testStore = lruStore
	b.Run("LRUon", benchWriteReceipts)

	testStore = simpleStore
	b.Run("LRUoff", benchWriteReceipts)
}

func benchWriteReceipts(b *testing.B) {
	expect := _createTestReceipts()

	if testStore.cache.Receipts != nil {
		testStore.cache.Receipts.Purge()
	}

	for i := 0; i < b.N; i++ {
		testStore.SetReceipts(1, *expect)
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
