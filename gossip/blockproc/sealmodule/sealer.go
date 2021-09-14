package sealmodule

import (
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/lachesis"

	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
)

type OperaEpochsSealerModule struct{}

func New() *OperaEpochsSealerModule {
	return &OperaEpochsSealerModule{}
}

func (m *OperaEpochsSealerModule) Start(block blockproc.BlockCtx) blockproc.SealerProcessor {
	return &OperaEpochsSealer{
		block: block,
	}
}

type OperaEpochsSealer struct {
	block blockproc.BlockCtx
}

func (s *OperaEpochsSealer) EpochSealing(bs blockproc.BlockState, es blockproc.EpochState) bool {
	sealEpoch := bs.EpochGas >= es.Rules.Epochs.MaxEpochGas
	sealEpoch = sealEpoch || (s.block.Time-es.EpochStart) >= es.Rules.Epochs.MaxEpochDuration
	sealEpoch = sealEpoch || bs.AdvanceEpochs > 0
	return sealEpoch || bs.EpochCheaters.Len() != 0
}

// SealEpoch is called after pre-internal transactions are executed
func (s *OperaEpochsSealer) SealEpoch(bs blockproc.BlockState, es blockproc.EpochState) (blockproc.BlockState, blockproc.EpochState) {
	// Select new validators
	oldValidators := es.Validators
	builder := pos.NewBigBuilder()
	for v, profile := range bs.NextValidatorProfiles {
		builder.Set(v, profile.Weight)
	}
	newValidators := builder.Build()
	es.Validators = newValidators
	es.ValidatorProfiles = bs.NextValidatorProfiles.Copy()

	// Build new []ValidatorEpochState and []ValidatorBlockState
	newValidatorEpochStates := make([]blockproc.ValidatorEpochState, newValidators.Len())
	newValidatorBlockStates := make([]blockproc.ValidatorBlockState, newValidators.Len())
	for newValIdx := idx.Validator(0); newValIdx < newValidators.Len(); newValIdx++ {
		// default values
		newValidatorBlockStates[newValIdx] = blockproc.ValidatorBlockState{
			Originated: new(big.Int),
		}
		// inherit validator's state from previous epoch, if any
		valID := newValidators.GetID(newValIdx)
		if !oldValidators.Exists(valID) {
			// new validator
			newValidatorBlockStates[newValIdx].LastBlock = s.block.Idx
			newValidatorBlockStates[newValIdx].LastOnlineTime = s.block.Time
			continue
		}
		oldValIdx := oldValidators.GetIdx(valID)
		newValidatorBlockStates[newValIdx] = bs.ValidatorStates[oldValIdx]
		newValidatorBlockStates[newValIdx].DirtyGasRefund = 0
		newValidatorBlockStates[newValIdx].Uptime = 0
		newValidatorEpochStates[newValIdx].GasRefund = bs.ValidatorStates[oldValIdx].DirtyGasRefund
		newValidatorEpochStates[newValIdx].PrevEpochEvent = bs.ValidatorStates[oldValIdx].LastEvent
	}
	es.ValidatorStates = newValidatorEpochStates
	bs.ValidatorStates = newValidatorBlockStates
	es.Validators = newValidators

	// dirty data becomes active
	es.PrevEpochStart = es.EpochStart
	es.EpochStart = s.block.Time
	es.Rules = bs.DirtyRules.Copy()
	es.EpochStateRoot = bs.FinalizedStateRoot

	bs.EpochGas = 0
	bs.EpochCheaters = lachesis.Cheaters{}
	newEpoch := es.Epoch + 1
	es.Epoch = newEpoch

	if bs.AdvanceEpochs > 0 {
		bs.AdvanceEpochs--
	}

	return bs, es
}
