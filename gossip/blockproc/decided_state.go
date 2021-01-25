package blockproc

import (
	"crypto/sha256"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/rlp"

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
	LastBlock             idx.Block
	LastCompleteStateRoot hash.Hash
	EpochGas              uint64

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

	EpochStateRoot hash.Hash

	Validators        *pos.Validators
	ValidatorStates   []ValidatorEpochState
	ValidatorProfiles ValidatorProfiles

	Rules opera.Rules
}

func (es *EpochState) GetValidatorState(id idx.ValidatorID, validators *pos.Validators) *ValidatorEpochState {
	validatorIdx := validators.GetIdx(id)
	return &es.ValidatorStates[validatorIdx]
}

func (es EpochState) Duration() inter.Timestamp {
	return es.EpochStart - es.PrevEpochStart
}

func (es EpochState) Hash() hash.Hash {
	hasher := sha256.New()
	err := rlp.Encode(hasher, &es)
	if err != nil {
		panic("can't hash: " + err.Error())
	}
	return hash.BytesToHash(hasher.Sum(nil))
}
