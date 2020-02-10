package app

/*
	In LRU cache data stored like value
*/

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

type receiptRLP struct {
	Receipt *types.ReceiptForStorage
	// These fields aren't serialized in types.ReceiptForStorage
	ContractAddress common.Address
	GasUsed         uint64
}

// SetReceipts stores transaction receipts.
func (s *Store) SetReceipts(n idx.Block, receipts types.Receipts) {
	receiptsStorage := make([]*receiptRLP, len(receipts))
	for i, r := range receipts {
		receiptsStorage[i] = &receiptRLP{
			Receipt:         (*types.ReceiptForStorage)(r),
			ContractAddress: r.ContractAddress,
			GasUsed:         r.GasUsed,
		}
	}
	s.set(s.table.Receipts, n.Bytes(), receiptsStorage)

	// Add to LRU cache.
	if s.cache.Receipts != nil {
		s.cache.Receipts.Add(n, receiptsStorage)
	}
}

// GetReceipts returns stored transaction receipts.
func (s *Store) GetReceipts(n idx.Block) types.Receipts {
	var receiptsStorage *[]*receiptRLP

	// Get data from LRU cache first.
	if s.cache.Receipts != nil {
		if c, ok := s.cache.Receipts.Get(n); ok {
			if receiptsStorage, ok = c.(*[]*receiptRLP); !ok {
				if cv, ok := c.([]*receiptRLP); ok {
					receiptsStorage = &cv
				}
			}
		}
	}

	if receiptsStorage == nil {
		receiptsStorage, _ = s.get(s.table.Receipts, n.Bytes(), &[]*receiptRLP{}).(*[]*receiptRLP)
		if receiptsStorage == nil {
			return nil
		}

		// Add to LRU cache.
		if s.cache.Receipts != nil {
			s.cache.Receipts.Add(n, *receiptsStorage)
		}
	}

	receipts := make(types.Receipts, len(*receiptsStorage))
	for i, r := range *receiptsStorage {
		receipts[i] = (*types.Receipt)(r.Receipt)
		receipts[i].ContractAddress = r.ContractAddress
		receipts[i].GasUsed = r.GasUsed
	}
	return receipts
}
