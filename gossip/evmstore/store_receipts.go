package evmstore

/*
	In LRU cache data stored like value
*/

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"
)

// SetReceipts stores transaction receipts.
func (s *Store) SetReceipts(n idx.Block, receipts types.Receipts) {
	receiptsStorage := make([]*types.ReceiptForStorage, len(receipts))
	for i, r := range receipts {
		receiptsStorage[i] = (*types.ReceiptForStorage)(r)
	}
	s.SetRawReceipts(n, receiptsStorage)

	// Add to LRU cache.
	s.cache.Receipts.Add(n, receipts)
}

// SetRawReceipts stores raw transaction receipts.
func (s *Store) SetRawReceipts(n idx.Block, receipts []*types.ReceiptForStorage) {
	s.rlp.Set(s.table.Receipts, n.Bytes(), receipts)
	// Remove from LRU cache.
	s.cache.Receipts.Remove(n)
}

// GetReceipts returns stored transaction receipts.
func (s *Store) GetReceipts(n idx.Block) types.Receipts {
	var receiptsStorage *[]*types.ReceiptForStorage

	// Get data from LRU cache first.
	if s.cache.Receipts != nil {
		if c, ok := s.cache.Receipts.Get(n); ok {
			return c.(types.Receipts)
		}
	}

	receiptsStorage, _ = s.rlp.Get(s.table.Receipts, n.Bytes(), &[]*types.ReceiptForStorage{}).(*[]*types.ReceiptForStorage)
	if receiptsStorage == nil {
		return nil
	}

	receipts := make(types.Receipts, len(*receiptsStorage))
	for i, r := range *receiptsStorage {
		receipts[i] = (*types.Receipt)(r)
		var prev uint64
		if i != 0 {
			prev = receipts[i-1].CumulativeGasUsed
		}
		receipts[i].GasUsed = receipts[i].CumulativeGasUsed - prev
	}

	// Add to LRU cache.
	s.cache.Receipts.Add(n, receipts)

	return receipts
}
