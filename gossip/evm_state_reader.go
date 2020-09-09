package gossip

import (
	"sync"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/app"
	"github.com/Fantom-foundation/go-opera/evmcore"
)

type EvmStateReader struct {
	*ServiceFeed
	engineMu *sync.RWMutex

	store *Store
	app   *app.Store
}

func (s *Service) GetEvmStateReader() *EvmStateReader {
	return &EvmStateReader{
		ServiceFeed: &s.feed,
		engineMu:    s.engineMu,
		store:       s.store,
		app:         s.store.app,
	}
}

func (r *EvmStateReader) CurrentBlock() *evmcore.EvmBlock {
	n := r.store.GetLatestBlockIndex()

	return r.getBlock(hash.Event{}, n, n != 0)
}

func (r *EvmStateReader) CurrentHeader() *evmcore.EvmHeader {
	n := r.store.GetLatestBlockIndex()

	return r.getBlock(hash.Event{}, n, false).Header()
}

func (r *EvmStateReader) GetHeader(h common.Hash, n uint64) *evmcore.EvmHeader {
	return r.GetDagHeader(hash.Event(h), idx.Block(n))
}

func (r *EvmStateReader) GetBlock(h common.Hash, n uint64) *evmcore.EvmBlock {
	return r.GetDagBlock(hash.Event(h), idx.Block(n))
}

func (r *EvmStateReader) GetDagHeader(h hash.Event, n idx.Block) *evmcore.EvmHeader {
	return r.getBlock(h, n, false).Header()
}

func (r *EvmStateReader) GetDagBlock(h hash.Event, n idx.Block) *evmcore.EvmBlock {
	return r.getBlock(h, n, n != 0)
}

func (r *EvmStateReader) getBlock(h hash.Event, n idx.Block, readTxs bool) *evmcore.EvmBlock {
	block := r.store.GetBlock(n)
	if block == nil {
		return nil
	}
	if (h != hash.Event{}) && (h != block.Atropos) {
		return nil
	}

	transactions := make(types.Transactions, 0, len(block.Events)*10)
	if readTxs {
		txCount := uint32(0)
		skipCount := 0
		for _, id := range block.Events {
			e := r.store.GetEventPayload(id)
			if e == nil {
				log.Crit("Event not found", "event", id.String())
				continue
			}

			// appends txs except skipped ones
			for _, tx := range e.Txs() {
				if skipCount < len(block.SkippedTxs) && block.SkippedTxs[skipCount] == txCount {
					skipCount++
				} else {
					transactions = append(transactions, tx)
				}
				txCount++
			}
		}
	}

	var prev hash.Event
	if n != 0 {
		prev = r.store.GetBlock(n - 1).Atropos
	}
	evmHeader := evmcore.ToEvmHeader(block, n, prev)
	evmBlock := &evmcore.EvmBlock{
		EvmHeader: *evmHeader,
	}

	if readTxs {
		evmBlock.Transactions = transactions
	} else {
		evmBlock.Transactions = make(types.Transactions, 0)
	}

	return evmBlock
}

func (r *EvmStateReader) StateAt(root common.Hash) (*state.StateDB, error) {
	return r.app.StateDB(hash.Hash(root)), nil
}
