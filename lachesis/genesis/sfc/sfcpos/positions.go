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
	// Topics of SFC contract logs
	Topics = struct {
		CreatedStake                  common.Hash
		IncreasedStake                common.Hash
		CreatedDelegation             common.Hash
		PreparedToWithdrawStake       common.Hash
		PreparedToWithdrawDelegation  common.Hash
		WithdrawnStake                common.Hash
		WithdrawnDelegation           common.Hash
		ClaimedDelegationReward       common.Hash
		ClaimedValidatorReward        common.Hash
		UpdatedBaseRewardPerSec       common.Hash
		UpdatedGasPowerAllocationRate common.Hash
		DeactivatedStake              common.Hash
		DeactivatedDelegation         common.Hash
		UpdatedStake                  common.Hash
		UpdatedDelegation             common.Hash
	}{
		CreatedStake:                  hash.Of([]byte("CreatedStake(uint256,address,uint256)")),
		IncreasedStake:                hash.Of([]byte("IncreasedStake(uint256,uint256,uint256)")),
		CreatedDelegation:             hash.Of([]byte("CreatedDelegation(address,uint256,uint256)")),
		PreparedToWithdrawStake:       hash.Of([]byte("PreparedToWithdrawStake(uint256)")),
		PreparedToWithdrawDelegation:  hash.Of([]byte("PreparedToWithdrawDelegation(address, uint256)")),
		WithdrawnStake:                hash.Of([]byte("WithdrawnStake(uint256,uint256)")),
		WithdrawnDelegation:           hash.Of([]byte("WithdrawnDelegation(address,uint256,uint256)")),
		ClaimedDelegationReward:       hash.Of([]byte("ClaimedDelegationReward(address,uint256,uint256,uint256,uint256)")),
		ClaimedValidatorReward:        hash.Of([]byte("ClaimedValidatorReward(uint256,uint256,uint256,uint256)")),
		UpdatedBaseRewardPerSec:       hash.Of([]byte("UpdatedBaseRewardPerSec(uint256)")),
		UpdatedGasPowerAllocationRate: hash.Of([]byte("UpdatedGasPowerAllocationRate(uint256,uint256)")),
		DeactivatedStake:              hash.Of([]byte("DeactivatedStake(uint256)")),
		DeactivatedDelegation:         hash.Of([]byte("DeactivatedDelegation(address,uint256)")),
		UpdatedStake:                  hash.Of([]byte("UpdatedStake(uint256,uint256,uint256)")),
		UpdatedDelegation:             hash.Of([]byte("UpdatedDelegation(address,uint256,uint256,uint256)")),
	}
)

// Global variables

func Owner() common.Hash {
	return utils.U64to256(0)
}

const (
	offset = 30
)

func CurrentSealedEpoch() common.Hash {
	return utils.U64to256(offset + 0)
}

func StakersLastID() common.Hash {
	return utils.U64to256(offset + 4)
}

func StakersNum() common.Hash {
	return utils.U64to256(offset + 5)
}

func StakeTotalAmount() common.Hash {
	return utils.U64to256(offset + 6)
}

// Stake

type StakePos struct {
	object
}

func Staker(stakerID idx.StakerID) StakePos {
	position := getMapValue(common.Hash{}, utils.U64to256(uint64(stakerID)), offset+2)

	return StakePos{object{base: position.Big()}}
}

func (p *StakePos) Status() common.Hash {
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
	return getMapValue(common.Hash{}, vstaker.Hash(), offset+3)
}

// EpochSnapshot

type EpochSnapshotPos struct {
	object
}

func EpochSnapshot(epoch idx.Epoch) EpochSnapshotPos {
	position := getMapValue(common.Hash{}, utils.U64to256(uint64(epoch)), offset+1)

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

func (p *EpochSnapshotPos) TotalBaseRewardWeight() common.Hash {
	return p.Field(4)
}

func (p *EpochSnapshotPos) TotalTxRewardWeight() common.Hash {
	return p.Field(5)
}

func (p *EpochSnapshotPos) BaseRewardPerSecond() common.Hash {
	return p.Field(6)
}

func (p *EpochSnapshotPos) StakeTotalAmount() common.Hash {
	return p.Field(7)
}

func (p *EpochSnapshotPos) DelegationsTotalAmount() common.Hash {
	return p.Field(8)
}

func (p *EpochSnapshotPos) TotalSupply() common.Hash {
	return p.Field(9)
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

func (p *ValidatorMeritPos) StakeAmount() common.Hash {
	return p.Field(0)
}

func (p *ValidatorMeritPos) DelegatedMe() common.Hash {
	return p.Field(1)
}

func (p *ValidatorMeritPos) BaseRewardWeight() common.Hash {
	return p.Field(2)
}

func (p *ValidatorMeritPos) TxRewardWeight() common.Hash {
	return p.Field(3)
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
