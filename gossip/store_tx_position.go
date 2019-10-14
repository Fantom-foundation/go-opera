package gossip

/*
	In LRU cache data stored like pointer
*/

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

type TxPosition struct {
	Block       idx.Block
	Event       hash.Event
	EventOffset uint32
	BlockOffset uint32
}

// SetTxPosition stores transaction block and position.
func (s *Store) SetTxPosition(txid common.Hash, position *TxPosition) {
	s.set(s.table.TxPositions, txid.Bytes(), position)

	// Add to LRU cache.
	if position != nil && s.cache.TxPositions != nil {
		s.cache.TxPositions.Add(txid.String(), position)
	}
}

// GetTxPosition returns stored transaction block and position.
func (s *Store) GetTxPosition(txid common.Hash) *TxPosition {
	// Get data from LRU cache first.
	if s.cache.TxPositions != nil {
		if c, ok := s.cache.TxPositions.Get(txid.String()); ok {
			if b, ok := c.(*TxPosition); ok {
				return b
			}
		}
	}

	txPosition, _ := s.get(s.table.TxPositions, txid.Bytes(), &TxPosition{}).(*TxPosition)

	// Add to LRU cache.
	if txPosition != nil && s.cache.TxPositions != nil {
		s.cache.TxPositions.Add(txid.String(), txPosition)
	}

	return txPosition
}
