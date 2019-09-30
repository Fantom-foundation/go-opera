package gossip

import (
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// SetReceipt stores transaction receipts.
func (s *Store) SetReceipts(n idx.Block, receipts types.Receipts) {
	s.set(s.table.Receipts, n.Bytes(), receipts)
}

// GetReceipt returns stored transaction receipts.
func (s *Store) GetReceipts(n idx.Block) types.Receipts {
	receipts, _ := s.get(s.table.Receipts, n.Bytes(), &types.Receipts{}).(*types.Receipts)
	return *receipts
}
