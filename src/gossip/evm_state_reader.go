package gossip

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-lachesis/src/evm_core"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

type EvmStateReader struct {
	*ServiceFeed
	engineMu *sync.RWMutex
	engine   Consensus

	store *Store
}

func (s *Service) GetEvmStateReader() *EvmStateReader {
	return &EvmStateReader{
		ServiceFeed: &s.feed,
		engineMu:    s.engineMu,
		engine:      s.engine,
		store:       s.store,
	}
}

func (r *EvmStateReader) CurrentBlock() *evm_core.EvmBlock {
	r.engineMu.RLock()
	defer r.engineMu.RUnlock()

	n, h := r.engine.LastBlock()
	return r.getBlock(hash.Event(h), idx.Block(n), n != 0)
}

func (r *EvmStateReader) CurrentHeader() *evm_core.EvmHeader {
	r.engineMu.RLock()
	defer r.engineMu.RUnlock()

	n, h := r.engine.LastBlock()
	return r.getBlock(hash.Event(h), idx.Block(n), false).Header()
}

func (r *EvmStateReader) GetHeader(h common.Hash, n uint64) *evm_core.EvmHeader {
	return r.GetDagHeader(hash.Event(h), idx.Block(n))
}

func (r *EvmStateReader) GetBlock(h common.Hash, n uint64) *evm_core.EvmBlock {
	return r.GetDagBlock(hash.Event(h), idx.Block(n))
}

func (r *EvmStateReader) GetDagHeader(h hash.Event, n idx.Block) *evm_core.EvmHeader {
	r.engineMu.RLock()
	defer r.engineMu.RUnlock()

	return r.getBlock(h, n, false).Header()
}

func (r *EvmStateReader) GetDagBlock(h hash.Event, n idx.Block) *evm_core.EvmBlock {
	r.engineMu.RLock()
	defer r.engineMu.RUnlock()

	return r.getBlock(h, n, n != 0)
}

func (r *EvmStateReader) getBlock(h hash.Event, n idx.Block, readTxs bool) *evm_core.EvmBlock {
	block := r.store.GetBlock(n)
	if block == nil {
		return nil
	}
	if (h != hash.Event{}) && (h != block.Hash()) {
		return nil
	}

	evmHeader := evm_core.ToEvmHeader(block)
	evmBlock := &evm_core.EvmBlock{
		EvmHeader: *evmHeader,
	}

	if readTxs {
		evmBlock.Transactions = make(types.Transactions, 0, len(block.Events)*10)
		txCount := uint(0)
		skipCount := 0
		for _, id := range block.Events {
			e := r.store.GetEvent(id)
			if e == nil {
				log.Crit("Event wasn't found", "event", id.String())
				continue
			}

			// appends txs except skipped ones
			for _, tx := range e.Transactions {
				if skipCount < len(block.SkippedTxs) && block.SkippedTxs[skipCount] == txCount {
					skipCount++
				} else {
					evmBlock.Transactions = append(evmBlock.Transactions, tx)
				}
				txCount++
			}
		}
	} else {
		evmBlock.Transactions = make(types.Transactions, 0)
	}
	return evmBlock
}

func (r *EvmStateReader) StateAt(root common.Hash) (*state.StateDB, error) {
	return r.store.StateDB(root), nil
}
