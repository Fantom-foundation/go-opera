package blockproc

import (
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
)

type ValidatorBlockState struct {
	Cheater          bool
	LastEvent        hash.Event
	Uptime           inter.Timestamp
	LastMedianTime   inter.Timestamp
	LastGasPowerLeft inter.GasPowerLeft
	LastBlock        idx.Block
	DirtyGasRefund   uint64
	Originated       *big.Int
}

var DefaultValidatorBlockState = ValidatorBlockState{
	Originated: new(big.Int),
}

type ValidatorEpochState struct {
	GasRefund      uint64
	PrevEpochEvent hash.Event
}

type BlockState struct {
	LastBlock     idx.Block
	LastStateRoot hash.Hash
	EpochGas      uint64

	ValidatorStates       []ValidatorBlockState
	NextValidatorProfiles ValidatorProfiles

	DirtyRules opera.Rules

	AdvanceEpochs idx.Epoch
}

func (bs *BlockState) GetValidatorState(id idx.ValidatorID, validators *pos.Validators) *ValidatorBlockState {
	validatorIdx := validators.GetIdx(id)
	return &bs.ValidatorStates[validatorIdx]
}

type EpochState struct {
	Epoch          idx.Epoch
	EpochStart     inter.Timestamp
	PrevEpochStart inter.Timestamp

	Validators        *pos.Validators
	ValidatorStates   []ValidatorEpochState
	ValidatorProfiles ValidatorProfiles

	Rules opera.Rules
}

func (es *EpochState) GetValidatorState(id idx.ValidatorID, validators *pos.Validators) *ValidatorEpochState {
	validatorIdx := validators.GetIdx(id)
	return &es.ValidatorStates[validatorIdx]
}

func (es *EpochState) Duration() inter.Timestamp {
	return es.EpochStart - es.PrevEpochStart
}
