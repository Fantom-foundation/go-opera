package sfcapi

import (
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

// GetDelegationClaimedRewards returns sum of claimed rewards in past, by this delegation
func (s *Store) GetDelegationClaimedRewards(id DelegationID) *big.Int {
	amount, err := s.table.DelegationOldRewards.Get(id.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key-value", "err", err)
	}
	if amount == nil {
		return big.NewInt(0)
	}
	return new(big.Int).SetBytes(amount)
}

// SetDelegationClaimedRewards sets sum of claimed rewards in past
func (s *Store) SetDelegationClaimedRewards(id DelegationID, amount *big.Int) {
	err := s.table.DelegationOldRewards.Put(id.Bytes(), amount.Bytes())
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// IncDelegationClaimedRewards increments sum of claimed rewards in past
func (s *Store) IncDelegationClaimedRewards(id DelegationID, diff *big.Int) {
	amount := s.GetDelegationClaimedRewards(id)
	amount.Add(amount, diff)
	s.SetDelegationClaimedRewards(id, amount)
}

// GetStakerDelegationsClaimedRewards returns sum of claimed rewards in past, by this delegations of this staker
func (s *Store) GetStakerDelegationsClaimedRewards(stakerID idx.ValidatorID) *big.Int {
	amount, err := s.table.StakerDelegationsOldRewards.Get(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key-value", "err", err)
	}
	if amount == nil {
		return big.NewInt(0)
	}
	return new(big.Int).SetBytes(amount)
}

// SetStakerDelegationsClaimedRewards sets sum of claimed rewards in past
func (s *Store) SetStakerDelegationsClaimedRewards(stakerID idx.ValidatorID, amount *big.Int) {
	err := s.table.StakerDelegationsOldRewards.Put(stakerID.Bytes(), amount.Bytes())
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// IncStakerDelegationsClaimedRewards increments sum of claimed rewards in past
func (s *Store) IncStakerDelegationsClaimedRewards(stakerID idx.ValidatorID, diff *big.Int) {
	amount := s.GetStakerDelegationsClaimedRewards(stakerID)
	amount.Add(amount, diff)
	s.SetStakerDelegationsClaimedRewards(stakerID, amount)
}
