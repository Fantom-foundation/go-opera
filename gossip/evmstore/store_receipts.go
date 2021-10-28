package evmstore

/*
	In LRU cache data stored like value
*/

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// SetReceipts stores transaction receipts.
func (s *Store) SetReceipts(n idx.Block, receipts types.Receipts) {
	receiptsStorage := make([]*types.ReceiptForStorage, len(receipts))
	for i, r := range receipts {
		receiptsStorage[i] = (*types.ReceiptForStorage)(r)
	}

	size := s.SetRawReceipts(n, receiptsStorage)

	// Add to LRU cache.
	s.cache.Receipts.Add(n, receipts, uint(size))
}

// SetRawReceipts stores raw transaction receipts.
func (s *Store) SetRawReceipts(n idx.Block, receipts []*types.ReceiptForStorage) (size int) {
	buf, err := rlp.EncodeToBytes(receipts)
	if err != nil {
		s.Log.Crit("Failed to encode rlp", "err", err)
	}

	if err := s.table.Receipts.Put(n.Bytes(), buf); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}

	// Remove from LRU cache.
	s.cache.Receipts.Remove(n)

	return len(buf)
}

func (s *Store) GetRawReceiptsRLP(n idx.Block) rlp.RawValue {
	buf, err := s.table.Receipts.Get(n.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	return buf
}

func (s *Store) GetRawReceipts(n idx.Block) ([]*types.ReceiptForStorage, int) {
	buf := s.GetRawReceiptsRLP(n)
	if buf == nil {
		return nil, 0
	}

	var receiptsStorage []*types.ReceiptForStorage
	err := rlp.DecodeBytes(buf, &receiptsStorage)
	if err != nil {
		s.Log.Crit("Failed to decode rlp", "err", err, "size", len(buf))
	}
	return receiptsStorage, len(buf)
}

// GetReceipts returns stored transaction receipts.
func (s *Store) GetReceipts(n idx.Block) types.Receipts {
	// Get data from LRU cache first.
	if s.cache.Receipts != nil {
		if c, ok := s.cache.Receipts.Get(n); ok {
			return c.(types.Receipts)
		}
	}

	receiptsStorage, size := s.GetRawReceipts(n)

	receipts := make(types.Receipts, len(receiptsStorage))
	for i, r := range receiptsStorage {
		receipts[i] = (*types.Receipt)(r)
		var prev uint64
		if i != 0 {
			prev = receipts[i-1].CumulativeGasUsed
		}
		receipts[i].GasUsed = receipts[i].CumulativeGasUsed - prev
	}

	// Add to LRU cache.
	s.cache.Receipts.Add(n, receipts, uint(size))

	return receipts
}
