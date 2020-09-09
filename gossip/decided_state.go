package gossip

import (
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/sfctype"
)

type ValidatorBlockState struct {
	PrevEvent        hash.Event
	Uptime           inter.Timestamp
	PrevMedianTime   inter.Timestamp
	PrevGasPowerLeft inter.GasPowerLeft
	PrevBlock        idx.Block
	DirtyGasRefund   uint64
	Originated       *big.Int
}

type ValidatorEpochState struct {
	GasRefund      uint64
	PrevEpochEvent hash.Event
}

type BlockState struct {
	Block       idx.Block
	EpochBlocks idx.Block
	EpochFee    *big.Int

	ValidatorStates []ValidatorBlockState
}

func (bs *BlockState) GetValidatorState(id idx.ValidatorID, validators *pos.Validators) *ValidatorBlockState {
	validatorIdx := validators.GetIdx(id)
	return &bs.ValidatorStates[validatorIdx]
}

type EpochState struct {
	Epoch          idx.Epoch
	Validators     *pos.Validators
	EpochStart     inter.Timestamp
	PrevEpochStart inter.Timestamp

	ValidatorStates   []ValidatorEpochState
	ValidatorProfiles []sfctype.SfcStakerAndID
}

func (es *EpochState) GetValidatorState(id idx.ValidatorID, validators *pos.Validators) *ValidatorEpochState {
	validatorIdx := validators.GetIdx(id)
	return &es.ValidatorStates[validatorIdx]
}

func (es *EpochState) Duration() inter.Timestamp {
	return es.EpochStart - es.PrevEpochStart
}
