package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/evm_core"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"sync"
)

type EvmStateReader struct {
	*ServiceFeed
	engineMu *sync.RWMutex
	engine   Consensus

	store *Store
}

func (r *EvmStateReader) CurrentBlock() *evm_core.EvmBlock {
	r.engineMu.RLock()
	defer r.engineMu.RUnlock()

	n, h := r.engine.LastBlock()
	return r.getBlock(common.Hash(h), uint64(n))
}

func (r *EvmStateReader) GetBlock(h common.Hash, n uint64) *evm_core.EvmBlock {
	r.engineMu.RLock()
	defer r.engineMu.RUnlock()

	return r.getBlock(common.Hash(h), uint64(n))

}

func (r *EvmStateReader) getBlock(h common.Hash, n uint64) *evm_core.EvmBlock {
	block := r.store.GetBlock(idx.Block(n))
	if block.Hash() != hash.Event(h) {
		return nil
	}

	evm_header := evm_core.ToEvmHeader(block)
	evm_block := &evm_core.EvmBlock{
		EvmHeader:    *evm_header,
		Transactions: make(types.Transactions, 0, len(block.Events)*10),
	}
	for _, id := range block.Events {
		e := r.store.GetEvent(id)
		if e == nil {
			log.Fatal("Event wasn't found", "e", id.String())
		}

		evm_block.Transactions = append(evm_block.Transactions, e.Transactions...)
	}
	return evm_block

}

func (r *EvmStateReader) StateAt(root common.Hash) (*state.StateDB, error) {
	return r.store.StateDB(hash.Hash(root)), nil
}
