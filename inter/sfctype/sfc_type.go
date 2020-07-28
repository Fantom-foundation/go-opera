package sfctype

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

var (
	// ForkBit is set if staker has a confirmed pair of fork events
	ForkBit = uint64(1)
	// OfflineBit is set if staker has didn't have confirmed events for a long time
	OfflineBit = uint64(1 << 8)
	// CheaterMask is a combination of severe misbehavings
	CheaterMask = ForkBit
)

const (
	// DelegationIDSize is size of DelegationID serialized object
	DelegationIDSize = 20 + 4
)

// SfcStaker is the node-side representation of SFC staker
type SfcStaker struct {
	CreatedEpoch idx.Epoch
	CreatedTime  inter.Timestamp

	DeactivatedEpoch idx.Epoch
	DeactivatedTime  inter.Timestamp

	StakeAmount *big.Int
	DelegatedMe *big.Int

	Address common.Address

	Status uint64

	IsValidator bool `rlp:"-"` // API-only field
}

// Ok returns true if not deactivated and not pruned
func (s *SfcStaker) Ok() bool {
	return s.Status == 0 && s.DeactivatedEpoch == 0
}

// IsCheater returns true if staker is cheater
func (s *SfcStaker) IsCheater() bool {
	return s.Status&CheaterMask != 0
}

// HasFork returns true if staker has a confirmed fork
func (s *SfcStaker) HasFork() bool {
	return s.Status&ForkBit != 0
}

// Offline returns true if staker was offline for long time
func (s *SfcStaker) Offline() bool {
	return s.Status&OfflineBit != 0
}

// SfcStakerAndID is pair SfcStaker + StakerID
type SfcStakerAndID struct {
	StakerID idx.StakerID
	Staker   *SfcStaker
}

// CalcTotalStake returns sum of staker's stake and delegated to staker stake
func (s *SfcStaker) CalcTotalStake() *big.Int {
	return new(big.Int).Add(s.StakeAmount, s.DelegatedMe)
}

// SfcDelegation is the node-side representation of SFC delegation
type SfcDelegation struct {
	CreatedEpoch idx.Epoch
	CreatedTime  inter.Timestamp

	DeactivatedEpoch idx.Epoch
	DeactivatedTime  inter.Timestamp

	Amount *big.Int
}

// DelegationID is a pair of delegator address and staker ID to which delegation is applied
type DelegationID struct {
	Delegator common.Address
	StakerID  idx.StakerID
}

func (id *DelegationID) Bytes() []byte {
	return append(id.Delegator.Bytes(), id.StakerID.Bytes()...)
}

func BytesToDelegationID(bb []byte) DelegationID {
	if len(bb) < DelegationIDSize {
		panic("delegation ID deserialization failed")
	}
	delegator := common.BytesToAddress(bb[:20])
	stakerID := idx.BytesToStakerID(bb[20:])
	return DelegationID{delegator, stakerID}
}

// SfcDelegationAndID is pair SfcDelegation + address
type SfcDelegationAndID struct {
	Delegation *SfcDelegation
	ID         DelegationID
}

// EpochStats stores general statistics for an epoch
type EpochStats struct {
	Start    inter.Timestamp
	End      inter.Timestamp
	TotalFee *big.Int

	Epoch                 idx.Epoch `rlp:"-"` // API-only field
	TotalBaseRewardWeight *big.Int  `rlp:"-"` // API-only field
	TotalTxRewardWeight   *big.Int  `rlp:"-"` // API-only field
}

// Duration returns epoch duration
func (s *EpochStats) Duration() inter.Timestamp {
	return s.End - s.Start
}
