package gossip

import (
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/gasprice"
	"github.com/Fantom-foundation/go-opera/opera"
)

type EvmStateReader struct {
	*ServiceFeed

	store *Store
	gpo   *gasprice.Oracle
}

func NewEvmStateReader(s *Store) *EvmStateReader {
	return &EvmStateReader{
		store: s,
	}
}

func (s *Service) GetEvmStateReader() *EvmStateReader {
	return &EvmStateReader{
		ServiceFeed: &s.feed,
		store:       s.store,
		gpo:         s.gpo,
	}
}

// MinGasPrice returns current hard lower bound for gas price
func (r *EvmStateReader) MinGasPrice() *big.Int {
	return r.store.GetRules().Economy.MinGasPrice
}

// EffectiveMinTip returns current soft lower bound for gas tip
func (r *EvmStateReader) EffectiveMinTip() *big.Int {
	min := r.MinGasPrice()
	est := r.gpo.EffectiveMinGasPrice()
	est.Sub(est, min)
	if est.Sign() < 0 {
		return new(big.Int)
	}
	return est
}

func (r *EvmStateReader) MaxGasLimit() uint64 {
	rules := r.store.GetRules()
	maxEmptyEventGas := rules.Economy.Gas.EventGas +
		uint64(rules.Dag.MaxParents-rules.Dag.MaxFreeParents)*rules.Economy.Gas.ParentGas +
		uint64(rules.Dag.MaxExtraData)*rules.Economy.Gas.ExtraDataGas
	if rules.Economy.Gas.MaxEventGas < maxEmptyEventGas {
		return 0
	}
	return rules.Economy.Gas.MaxEventGas - maxEmptyEventGas
}

func (r *EvmStateReader) Config() *params.ChainConfig {
	return r.store.GetEvmChainConfig()
}

func (r *EvmStateReader) CurrentBlock() *evmcore.EvmBlock {
	n := r.store.GetLatestBlockIndex()

	return r.getBlock(hash.Event{}, n, true)
}

func (r *EvmStateReader) CurrentHeader() *evmcore.EvmHeader {
	n := r.store.GetLatestBlockIndex()

	return r.getBlock(hash.Event{}, n, false).Header()
}

func (r *EvmStateReader) GetHeader(h common.Hash, n uint64) *evmcore.EvmHeader {
	return r.getBlock(hash.Event(h), idx.Block(n), false).Header()
}

func (r *EvmStateReader) GetBlock(h common.Hash, n uint64) *evmcore.EvmBlock {
	return r.getBlock(hash.Event(h), idx.Block(n), true)
}

func (r *EvmStateReader) getBlock(h hash.Event, n idx.Block, readTxs bool) *evmcore.EvmBlock {
	block := r.store.GetBlock(n)
	if block == nil {
		return nil
	}
	if (h != hash.Event{}) && (h != block.Atropos) {
		return nil
	}
	if readTxs {
		if cached := r.store.EvmStore().GetCachedEvmBlock(n); cached != nil {
			return cached
		}
	}

	var transactions types.Transactions
	if readTxs {
		transactions = r.store.GetBlockTxs(n, block)
	} else {
		transactions = make(types.Transactions, 0)
	}

	// find block rules
	epoch := block.Atropos.Epoch()
	es := r.store.GetHistoryEpochState(epoch)
	var rules opera.Rules
	if es != nil {
		rules = es.Rules
	}
	var prev hash.Event
	if n != 0 {
		block := r.store.GetBlock(n - 1)
		if block != nil {
			prev = block.Atropos
		}
	}
	evmHeader := evmcore.ToEvmHeader(block, n, prev, rules)

	var evmBlock *evmcore.EvmBlock
	if readTxs {
		evmBlock = evmcore.NewEvmBlock(evmHeader, transactions)
		r.store.EvmStore().SetCachedEvmBlock(n, evmBlock)
	} else {
		// not completed block here
		evmBlock = &evmcore.EvmBlock{
			EvmHeader: *evmHeader,
		}
	}

	return evmBlock
}

func (r *EvmStateReader) StateAt(root common.Hash) (*state.StateDB, error) {
	return r.store.evm.StateDB(hash.Hash(root))
}
