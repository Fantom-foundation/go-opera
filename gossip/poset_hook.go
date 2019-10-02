package gossip

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/vector"
)

type HookedEngine struct {
	engine Consensus

	processEvent func(realEngine Consensus, e *inter.Event) error
}

// not safe for concurrent use
func (hook *HookedEngine) ProcessEvent(e *inter.Event) error {
	return hook.processEvent(hook.engine, e)
}

func (hook *HookedEngine) GetVectorIndex() *vector.Index {
	if hook.engine == nil {
		return nil
	}
	return hook.engine.GetVectorIndex()
}

func (hook *HookedEngine) GetGenesisHash() common.Hash {
	if hook.engine == nil {
		return common.Hash{}
	}
	return hook.engine.GetGenesisHash()
}

func (hook *HookedEngine) Prepare(e *inter.Event) *inter.Event {
	if hook.engine == nil {
		return e
	}
	return hook.engine.Prepare(e)
}

func (hook *HookedEngine) GetEpoch() idx.Epoch {
	if hook.engine == nil {
		return 1
	}
	return hook.engine.GetEpoch()
}

func (hook *HookedEngine) GetEpochValidators() (pos.Validators, idx.Epoch) {
	if hook.engine == nil {
		return pos.Validators{}, 1
	}
	return hook.engine.GetEpochValidators()
}

func (hook *HookedEngine) LastBlock() (idx.Block, hash.Event) {
	if hook.engine == nil {
		return idx.Block(1), hash.ZeroEvent
	}
	return hook.engine.LastBlock()
}

func (hook *HookedEngine) GetValidators() pos.Validators {
	if hook.engine == nil {
		return pos.Validators{}
	}
	return hook.engine.GetValidators()
}

func (hook *HookedEngine) GetConsensusTime(id hash.Event) (inter.Timestamp, error) {
	if hook.engine == nil {
		return 0, nil
	}
	return hook.engine.GetConsensusTime(id)
}

func (hook *HookedEngine) Bootstrap(fn inter.ApplyBlockFn) {
	if hook.engine == nil {
		return
	}
	hook.engine.Bootstrap(fn)
}
