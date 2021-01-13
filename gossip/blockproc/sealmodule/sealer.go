package sealmodule

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"

	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/inter"
)

type OperaEpochsSealerModule struct{}

func New() *OperaEpochsSealerModule {
	return &OperaEpochsSealerModule{}
}

func (m *OperaEpochsSealerModule) Start(block blockproc.BlockCtx, bs blockproc.BlockState, es blockproc.EpochState) blockproc.SealerProcessor {
	return &OperaEpochsSealer{
		block: block,
		es:    es,
		bs:    bs,
	}
}

type OperaEpochsSealer struct {
	block blockproc.BlockCtx
	es    blockproc.EpochState
	bs    blockproc.BlockState
}

func (s *OperaEpochsSealer) EpochSealing() bool {
	sealEpoch := s.bs.EpochGas >= s.es.Rules.Dag.MaxEpochGas
	sealEpoch = sealEpoch || (s.block.Time-s.es.EpochStart) >= s.es.Rules.Dag.MaxEpochDuration
	sealEpoch = sealEpoch || s.bs.AdvanceEpochs > 0
	return sealEpoch || s.block.CBlock.Cheaters.Len() != 0
}

func (p *OperaEpochsSealer) Update(bs blockproc.BlockState, es blockproc.EpochState) {
	p.bs, p.es = bs, es
}

func (s *OperaEpochsSealer) SealEpoch() (blockproc.BlockState, blockproc.EpochState) {
	// add final uptime for validators
	for _, info := range s.bs.ValidatorStates {
		if s.bs.LastBlock-info.LastBlock <= s.es.Rules.Economy.BlockMissedSlack {
			info.Uptime += inter.MaxTimestamp(s.block.Time, info.LastMedianTime) - info.LastMedianTime
		}
	}

	// Select new validators
	oldValidators := s.es.Validators
	builder := pos.NewBigBuilder()
	for v, profile := range s.bs.NextValidatorProfiles {
		builder.Set(v, profile.Weight)
	}
	newValidators := builder.Build()
	s.es.Validators = newValidators
	s.es.ValidatorProfiles = s.bs.NextValidatorProfiles

	// Build new []ValidatorEpochState and []ValidatorBlockState
	newValidatorEpochStates := make([]blockproc.ValidatorEpochState, newValidators.Len())
	newValidatorBlockStates := make([]blockproc.ValidatorBlockState, newValidators.Len())
	for newValIdx := idx.Validator(0); newValIdx < newValidators.Len(); newValIdx++ {
		// default values
		newValidatorBlockStates[newValIdx] = blockproc.DefaultValidatorBlockState
		// inherit validator's state from previous epoch, if any
		valID := newValidators.GetID(newValIdx)
		if !oldValidators.Exists(valID) {
			continue
		}
		oldValIdx := oldValidators.GetIdx(valID)
		newValidatorBlockStates[newValIdx] = s.bs.ValidatorStates[oldValIdx]
		newValidatorBlockStates[newValIdx].DirtyGasRefund = 0
		newValidatorBlockStates[newValIdx].Uptime = 0
		newValidatorEpochStates[newValIdx].GasRefund = s.bs.ValidatorStates[oldValIdx].DirtyGasRefund
		newValidatorEpochStates[newValIdx].PrevEpochEvent = s.bs.ValidatorStates[oldValIdx].LastEvent
	}
	s.es.ValidatorStates = newValidatorEpochStates
	s.bs.ValidatorStates = newValidatorBlockStates
	s.es.Validators = newValidators

	// dirty data becomes active
	s.es.PrevEpochStart = s.es.EpochStart
	s.es.EpochStart = s.block.Time
	s.es.Rules = s.bs.DirtyRules

	s.bs.EpochGas = 0
	newEpoch := s.es.Epoch + 1
	s.es.Epoch = newEpoch

	if s.bs.AdvanceEpochs > 0 {
		s.bs.AdvanceEpochs--
	}

	return s.bs, s.es
}
