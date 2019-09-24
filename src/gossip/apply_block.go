package gossip

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/src/evm_core"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

// ApplyBlock execs ordered txns on state.
func (s *Service) ApplyBlock(block *inter.Block, stateHash common.Hash, members pos.Members) (newStateHash common.Hash, newMembers pos.Members) {
	evmProcessor := evm_core.NewStateProcessor(params.AllEthashProtocolChanges, s.GetEvmStateReader())

	// Assemble block data
	evm_header := evm_core.ToEvmHeader(block)
	evmBlock := &evm_core.EvmBlock{
		EvmHeader:    *evm_header,
		Transactions: make(types.Transactions, 0, len(block.Events)*10),
	}
	for _, id := range block.Events {
		e := s.store.GetEvent(id)
		if e == nil {
			s.Log.Crit("Event wasn't found", "event", id.String())
		}

		evmBlock.Transactions = append(evmBlock.Transactions, e.Transactions...)
	}
	s.occurredTxs.CollectConfirmedTxs(evmBlock.Transactions) // TODO collect all the confirmed txs, not only block txs

	// Process txs
	statedb := s.store.StateDB(stateHash)
	_, _, _, totalFee, skipped, err := evmProcessor.Process(evmBlock, statedb, vm.Config{}, false)
	if err != nil {
		s.Log.Crit("Shouldn't happen ever because it's not strict", "err", err)
	}
	block.SkippedTxs = skipped

	// apply block rewards here if needed
	log.Info("New block", "index", block.Index, "hash", block.Hash().String(), "fee", totalFee, "txs", len(evmBlock.Transactions), "skipped_txs", len(skipped))

	// finalize
	newStateHash, err = statedb.Commit(true)
	if err != nil {
		s.Log.Crit("Failed to commit state", "err", err)
	}
	block.Root = newStateHash
	evmBlock.Root = newStateHash
	s.store.SetBlock(block)

	// new members
	// TODO replace with special transactions for changing members state
	// TODO the schema below doesn't work in all the cases, and intended only for testing
	{
		newMembers = members.Copy()
		for addr := range members {
			stake := pos.BalanceToStake(statedb.GetBalance(addr))
			newMembers.Set(addr, stake)
		}
		for _, tx := range evmBlock.Transactions {
			if tx.To() == nil {
				continue
			}
			stake := pos.BalanceToStake(statedb.GetBalance(*tx.To()))
			newMembers.Set(*tx.To(), stake)
		}
	}

	// filter skipped transactions before notifying
	skipCount := 0
	filteredTxs := make(types.Transactions, 0, len(evmBlock.Transactions))
	for i, tx := range evmBlock.Transactions {
		if skipCount < len(block.SkippedTxs) && block.SkippedTxs[skipCount] == uint(i) {
			skipCount++
		} else {
			filteredTxs = append(filteredTxs, tx)
		}
	}
	evmBlock.Transactions = filteredTxs

	s.feed.newBlock.Send(evm_core.ChainHeadNotify{evmBlock})

	return newStateHash, newMembers
}
