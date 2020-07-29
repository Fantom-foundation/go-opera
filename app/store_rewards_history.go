package app

import (
	"math/big"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/sfctype"
)

// GetStakerClaimedRewards returns sum of claimed rewards in past, by this staker
func (s *Store) GetStakerClaimedRewards(stakerID idx.StakerID) *big.Int {
	amount, err := s.table.StakerOldRewards.Get(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key-value", "err", err)
	}
	if amount == nil {
		return big.NewInt(0)
	}
	return new(big.Int).SetBytes(amount)
}

// SetStakerClaimedRewards sets sum of claimed rewards in past
func (s *Store) SetStakerClaimedRewards(stakerID idx.StakerID, amount *big.Int) {
	err := s.table.StakerOldRewards.Put(stakerID.Bytes(), amount.Bytes())
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// IncStakerClaimedRewards increments sum of claimed rewards in past
func (s *Store) IncStakerClaimedRewards(stakerID idx.StakerID, diff *big.Int) {
	amount := s.GetStakerClaimedRewards(stakerID)
	amount.Add(amount, diff)
	s.SetStakerClaimedRewards(stakerID, amount)
}

// DelStakerClaimedRewards deletes record about sum of claimed rewards in past
func (s *Store) DelStakerClaimedRewards(stakerID idx.StakerID) {
	err := s.table.StakerOldRewards.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key-value", "err", err)
	}
}

// GetDelegationClaimedRewards returns sum of claimed rewards in past, by this delegation
func (s *Store) GetDelegationClaimedRewards(id sfctype.DelegationID) *big.Int {
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
func (s *Store) SetDelegationClaimedRewards(id sfctype.DelegationID, amount *big.Int) {
	err := s.table.DelegationOldRewards.Put(id.Bytes(), amount.Bytes())
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// IncDelegationClaimedRewards increments sum of claimed rewards in past
func (s *Store) IncDelegationClaimedRewards(id sfctype.DelegationID, diff *big.Int) {
	amount := s.GetDelegationClaimedRewards(id)
	amount.Add(amount, diff)
	s.SetDelegationClaimedRewards(id, amount)
}

// DelDelegationClaimedRewards deletes record about sum of claimed rewards in past
func (s *Store) DelDelegationClaimedRewards(id sfctype.DelegationID) {
	err := s.table.DelegationOldRewards.Delete(id.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key-value", "err", err)
	}
}

// GetStakerDelegationsClaimedRewards returns sum of claimed rewards in past, by this delegations of this staker
func (s *Store) GetStakerDelegationsClaimedRewards(stakerID idx.StakerID) *big.Int {
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
func (s *Store) SetStakerDelegationsClaimedRewards(stakerID idx.StakerID, amount *big.Int) {
	err := s.table.StakerDelegationsOldRewards.Put(stakerID.Bytes(), amount.Bytes())
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// IncStakerDelegationsClaimedRewards increments sum of claimed rewards in past
func (s *Store) IncStakerDelegationsClaimedRewards(stakerID idx.StakerID, diff *big.Int) {
	amount := s.GetStakerDelegationsClaimedRewards(stakerID)
	amount.Add(amount, diff)
	s.SetStakerDelegationsClaimedRewards(stakerID, amount)
}

// DelStakerDelegationsClaimedRewards deletes record about sum of claimed rewards in past
func (s *Store) DelStakerDelegationsClaimedRewards(stakerID idx.StakerID) {
	err := s.table.StakerDelegationsOldRewards.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key-value", "err", err)
	}
}
