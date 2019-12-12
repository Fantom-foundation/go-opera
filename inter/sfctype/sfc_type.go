package sfctype

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
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

	IsCheater bool

	IsValidator bool `rlp:"-"` // API-only field
}

// SfcStakerAndID is pair SfcStaker + StakerID
type SfcStakerAndID struct {
	StakerID idx.StakerID
	Staker   *SfcStaker
}

// CalcTotalStake returns sum of staker's stake and delegated to staker stake
func (st *SfcStaker) CalcTotalStake() *big.Int {
	return new(big.Int).Add(st.StakeAmount, st.DelegatedMe)
}

// SfcDelegator is the node-side representation of SFC delegator
type SfcDelegator struct {
	CreatedEpoch idx.Epoch
	CreatedTime  inter.Timestamp

	DeactivatedEpoch idx.Epoch
	DeactivatedTime  inter.Timestamp

	Amount *big.Int

	ToStakerID idx.StakerID
}

// SfcDelegatorAndAddr is pair SfcDelegator + address
type SfcDelegatorAndAddr struct {
	Delegator *SfcDelegator
	Addr      common.Address
}

// EpochStats stores general statistics for an epoch
type EpochStats struct {
	Start    inter.Timestamp
	End      inter.Timestamp
	TotalFee *big.Int
}
