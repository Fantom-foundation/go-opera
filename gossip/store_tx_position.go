package gossip

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
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
}

// GetTxPosition returns stored transaction block and position.
func (s *Store) GetTxPosition(txid common.Hash) *TxPosition {
	txPosition, _ := s.get(s.table.TxPositions, txid.Bytes(), &TxPosition{}).(*TxPosition)
	return txPosition
}
