package evmstore

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/inter"
)

// SetTx stores non-event transaction.
func (s *Store) SetTx(txid common.Hash, tx *types.Transaction) {
	s.rlp.Set(s.table.Txs, txid.Bytes(), tx)
}

// GetTx returns stored non-event transaction.
func (s *Store) GetTx(txid common.Hash) *types.Transaction {
	tx, _ := s.rlp.Get(s.table.Txs, txid.Bytes(), &types.Transaction{}).(*types.Transaction)

	return tx
}

func (s *Store) GetBlockTxs(n idx.Block, block inter.Block, getEventPayload func(hash.Event) *inter.EventPayload) types.Transactions {
	if cached := s.GetCachedEvmBlock(n); cached != nil {
		return cached.Transactions
	}

	transactions := make(types.Transactions, 0, len(block.Txs)+len(block.InternalTxs)+len(block.Events)*10)
	for _, txid := range block.InternalTxs {
		tx := s.GetTx(txid)
		if tx == nil {
			log.Crit("Internal tx not found", "tx", txid.String())
			continue
		}
		transactions = append(transactions, tx)
	}
	for _, txid := range block.Txs {
		tx := s.GetTx(txid)
		if tx == nil {
			log.Crit("Tx not found", "tx", txid.String())
			continue
		}
		transactions = append(transactions, tx)
	}
	for _, id := range block.Events {
		e := getEventPayload(id)
		if e == nil {
			log.Crit("Block event not found", "event", id.String())
			continue
		}
		transactions = append(transactions, e.Txs()...)
	}

	transactions = inter.FilterSkippedTxs(transactions, block.SkippedTxs)

	return transactions
}
