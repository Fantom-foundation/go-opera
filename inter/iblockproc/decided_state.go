package iblockproc

import (
	"crypto/sha256"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
)

type ValidatorBlockState struct {
	LastEvent        EventInfo
	Uptime           inter.Timestamp
	LastOnlineTime   inter.Timestamp
	LastGasPowerLeft inter.GasPowerLeft
	LastBlock        idx.Block
	DirtyGasRefund   uint64
	Originated       *big.Int
}

type EventInfo struct {
	ID           hash.Event
	GasPowerLeft inter.GasPowerLeft
	Time         inter.Timestamp
}

type ValidatorEpochState struct {
	GasRefund      uint64
	PrevEpochEvent EventInfo
}

type BlockCtx struct {
	Idx     idx.Block
	Time    inter.Timestamp
	Atropos hash.Event
}

type BlockState struct {
	LastBlock          BlockCtx
	FinalizedStateRoot hash.Hash

	EpochGas        uint64
	EpochCheaters   lachesis.Cheaters
	CheatersWritten uint32

	ValidatorStates       []ValidatorBlockState
	NextValidatorProfiles inter.ValidatorProfiles

	DirtyRules *opera.Rules `rlp:"nil"` // nil means that there's no changes compared to epoch rules

	AdvanceEpochs idx.Epoch
}

func (bs BlockState) Copy() BlockState {
	cp := bs
	cp.EpochCheaters = make(lachesis.Cheaters, len(bs.EpochCheaters))
	copy(cp.EpochCheaters, bs.EpochCheaters)
	cp.ValidatorStates = make([]ValidatorBlockState, len(bs.ValidatorStates))
	copy(cp.ValidatorStates, bs.ValidatorStates)
	for i := range cp.ValidatorStates {
		cp.ValidatorStates[i].Originated = new(big.Int).Set(cp.ValidatorStates[i].Originated)
	}
	cp.NextValidatorProfiles = bs.NextValidatorProfiles.Copy()
	if bs.DirtyRules != nil {
		rules := bs.DirtyRules.Copy()
		cp.DirtyRules = &rules
	}
	return cp
}

func (bs *BlockState) GetValidatorState(id idx.ValidatorID, validators *pos.Validators) *ValidatorBlockState {
	validatorIdx := validators.GetIdx(id)
	return &bs.ValidatorStates[validatorIdx]
}

func (bs BlockState) Hash() hash.Hash {
	hasher := sha256.New()
	err := rlp.Encode(hasher, &bs)
	if err != nil {
		panic("can't hash: " + err.Error())
	}
	return hash.BytesToHash(hasher.Sum(nil))
}

type EpochStateV1 struct {
	Epoch          idx.Epoch
	EpochStart     inter.Timestamp
	PrevEpochStart inter.Timestamp

	EpochStateRoot hash.Hash

	Validators        *pos.Validators
	ValidatorStates   []ValidatorEpochState
	ValidatorProfiles inter.ValidatorProfiles

	Rules opera.Rules
}

type EpochState EpochStateV1

func (es *EpochState) GetValidatorState(id idx.ValidatorID, validators *pos.Validators) *ValidatorEpochState {
	validatorIdx := validators.GetIdx(id)
	return &es.ValidatorStates[validatorIdx]
}

func (es EpochState) Duration() inter.Timestamp {
	return es.EpochStart - es.PrevEpochStart
}

func (es EpochState) Hash() hash.Hash {
	var hashed interface{}
	if es.Rules.Upgrades.London {
		hashed = &es
	} else {
		es0 := EpochStateV0{
			Epoch:             es.Epoch,
			EpochStart:        es.EpochStart,
			PrevEpochStart:    es.PrevEpochStart,
			EpochStateRoot:    es.EpochStateRoot,
			Validators:        es.Validators,
			ValidatorStates:   make([]ValidatorEpochStateV0, len(es.ValidatorStates)),
			ValidatorProfiles: es.ValidatorProfiles,
			Rules:             es.Rules,
		}
		for i, v := range es.ValidatorStates {
			es0.ValidatorStates[i].GasRefund = v.GasRefund
			es0.ValidatorStates[i].PrevEpochEvent = v.PrevEpochEvent.ID
		}
		hashed = &es0
	}
	hasher := sha256.New()
	err := rlp.Encode(hasher, hashed)
	if err != nil {
		panic("can't hash: " + err.Error())
	}
	return hash.BytesToHash(hasher.Sum(nil))
}

func (es EpochState) Copy() EpochState {
	cp := es
	cp.ValidatorStates = make([]ValidatorEpochState, len(es.ValidatorStates))
	copy(cp.ValidatorStates, es.ValidatorStates)
	cp.ValidatorProfiles = es.ValidatorProfiles.Copy()
	if es.Rules != (opera.Rules{}) {
		cp.Rules = es.Rules.Copy()
	}
	return cp
}
