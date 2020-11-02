package evmstore

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// SetTxPosition stores transaction block and position.
func (s *Store) SetTx(txid common.Hash, tx *types.Transaction) {
	s.set(s.table.Txs, txid.Bytes(), tx)
}

// GetTxPosition returns stored transaction block and position.
func (s *Store) GetTx(txid common.Hash) *types.Transaction {
	tx, _ := s.get(s.table.Txs, txid.Bytes(), &types.Transaction{}).(*types.Transaction)

	return tx
}
