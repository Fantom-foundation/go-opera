package gossip

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/vector"
)

// HookedEngine is a wrapper around any engine, which hooks ProcessEvent()
type HookedEngine struct {
	engine Consensus

	processEvent func(realEngine Consensus, e *inter.Event) error
}

// ProcessEvent takes event into processing.
// Event order matter: parents first.
// ProcessEvent is not safe for concurrent use
func (hook *HookedEngine) ProcessEvent(e *inter.Event) error {
	return hook.processEvent(hook.engine, e)
}

// GetVectorIndex returns vector clock.
func (hook *HookedEngine) GetVectorIndex() *vector.Index {
	if hook.engine == nil {
		return nil
	}
	return hook.engine.GetVectorIndex()
}

// GetGenesisHash returns PrevEpochHash of first epoch.
func (hook *HookedEngine) GetGenesisHash() common.Hash {
	if hook.engine == nil {
		return common.Hash{}
	}
	return hook.engine.GetGenesisHash()
}

// Prepare fills consensus-related fields: Frame, IsRoot, MedianTimestamp, PrevEpochHash, GasPowerLeft
// returns nil if event should be dropped
func (hook *HookedEngine) Prepare(e *inter.Event) *inter.Event {
	if hook.engine == nil {
		return e
	}
	return hook.engine.Prepare(e)
}

// GetEpoch returns current epoch num to 3rd party.
func (hook *HookedEngine) GetEpoch() idx.Epoch {
	if hook.engine == nil {
		return 1
	}
	return hook.engine.GetEpoch()
}

// GetEpochValidators atomically returns validators of current epoch, and the epoch.
func (hook *HookedEngine) GetEpochValidators() (*pos.Validators, idx.Epoch) {
	if hook.engine == nil {
		return pos.EmptyValidators, 1
	}
	return hook.engine.GetEpochValidators()
}

// LastBlock returns current block.
func (hook *HookedEngine) LastBlock() (idx.Block, hash.Event) {
	if hook.engine == nil {
		return idx.Block(1), hash.ZeroEvent
	}
	return hook.engine.LastBlock()
}

// GetValidators returns validators of current epoch.
func (hook *HookedEngine) GetValidators() *pos.Validators {
	if hook.engine == nil {
		return pos.EmptyValidators
	}
	return hook.engine.GetValidators()
}

// GetConsensusTime calc consensus timestamp for given event, if event is confirmed.
func (hook *HookedEngine) GetConsensusTime(id hash.Event) (inter.Timestamp, error) {
	if hook.engine == nil {
		return 0, nil
	}
	return hook.engine.GetConsensusTime(id)
}

// Bootstrap restores poset's state from store.
func (hook *HookedEngine) Bootstrap(callbacks inter.ConsensusCallbacks) {
	if hook.engine == nil {
		return
	}
	hook.engine.Bootstrap(callbacks)
}
