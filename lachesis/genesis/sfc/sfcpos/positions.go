package sfcpos

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/utils"
)

// Events
var (
	CreateStakeTopic          = hash.Of([]byte("CreatedStake(uint256,address,uint256)"))
	IncreasedStakeTopic       = hash.Of([]byte("IncreasedStake(uint256,uint256,uint256)"))
	CreatedDelegationTopic    = hash.Of([]byte("CreatedDelegation(address,uint256,uint256)"))
	DeactivateStakeTopic      = hash.Of([]byte("PreparedToWithdrawStake(uint256)"))
	DeactivateDelegationTopic = hash.Of([]byte("PreparedToWithdrawDelegation(address)"))
)

// Global variables

func CurrentSealedEpoch() common.Hash {
	return utils.U64to256(0)
}

func StakersLastIdx() common.Hash {
	return utils.U64to256(4)
}

func StakersNum() common.Hash {
	return utils.U64to256(5)
}

func StakeTotalAmount() common.Hash {
	return utils.U64to256(6)
}

// Stake

type StakePos struct {
	object
}

func Staker(stakerID idx.StakerID) StakePos {
	position := getMapValue(common.Hash{}, utils.U64to256(uint64(stakerID)), 2)

	return StakePos{object{base: position.Big()}}
}

func (p *StakePos) IsCheater() common.Hash {
	return p.Field(0)
}

func (p *StakePos) CreatedEpoch() common.Hash {
	return p.Field(1)
}

func (p *StakePos) CreatedTime() common.Hash {
	return p.Field(2)
}

func (p *StakePos) StakeAmount() common.Hash {
	return p.Field(5)
}

func (p *StakePos) Address() common.Hash {
	return p.Field(8)
}

// stakerIDs

func StakerID(vstaker common.Address) common.Hash {
	return getMapValue(common.Hash{}, vstaker.Hash(), 3)
}

// EpochSnapshot

type EpochSnapshotPos struct {
	object
}

func EpochSnapshot(epoch idx.Epoch) EpochSnapshotPos {
	position := getMapValue(common.Hash{}, utils.U64to256(uint64(epoch)), 1)

	return EpochSnapshotPos{object{base: position.Big()}}
}

func (p *EpochSnapshotPos) EndTime() common.Hash {
	return p.Field(1)
}

func (p *EpochSnapshotPos) Duration() common.Hash {
	return p.Field(2)
}

func (p *EpochSnapshotPos) EpochFee() common.Hash {
	return p.Field(3)
}

func (p *EpochSnapshotPos) TotalValidatingPower() common.Hash {
	return p.Field(4)
}

// ValidatorMerit

type ValidatorMeritPos struct {
	object
}

func (p *EpochSnapshotPos) ValidatorMerit(stakerID idx.StakerID) ValidatorMeritPos {
	base := p.Field(0)

	position := getMapValue(base, utils.U64to256(uint64(stakerID)), 0)

	return ValidatorMeritPos{object{base: position.Big()}}
}

func (p *ValidatorMeritPos) ValidatingPower() common.Hash {
	return p.Field(0)
}

func (p *ValidatorMeritPos) StakeAmount() common.Hash {
	return p.Field(1)
}

func (p *ValidatorMeritPos) DelegatedMe() common.Hash {
	return p.Field(2)
}

// Util

func getMapValue(base common.Hash, key common.Hash, mapIdx int64) common.Hash {
	hasher := sha3.NewLegacyKeccak256()
	_, _ = hasher.Write(key.Bytes())
	start := base.Big()
	_, _ = hasher.Write(utils.BigTo256(start.Add(start, big.NewInt(mapIdx))).Bytes())

	return common.BytesToHash(hasher.Sum(nil))
}

type object struct {
	base *big.Int
}

func (p *object) Field(offset int64) common.Hash {
	if offset == 0 {
		return common.BytesToHash(p.base.Bytes())
	}

	start := new(big.Int).Set(p.base)

	return utils.BigTo256(start.Add(start, big.NewInt(offset)))
}
