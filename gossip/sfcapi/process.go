package sfcapi

import (
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera/contracts/sfc"
	"github.com/Fantom-foundation/go-opera/topicsdb"
)

func ApplyGenesis(s *Store, index *topicsdb.Index) {
	_ = index.ForEach(nil, [][]common.Hash{{sfc.ContractAddress.Hash()}, {Topics.ClaimedValidatorReward, Topics.ClaimedDelegationReward}}, func(l *types.Log) (gonext bool) {
		if l.Topics[0] == Topics.ClaimedValidatorReward && len(l.Topics) > 1 && len(l.Data) >= 32 {
			stakerID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
			reward := new(big.Int).SetBytes(l.Data[0:32])

			staker := s.GetSfcStaker(stakerID)
			if staker == nil {
				return true
			}
			s.IncDelegationClaimedRewards(DelegationID{staker.Address, stakerID}, reward)
			s.IncStakerDelegationsClaimedRewards(stakerID, reward)
		} else if l.Topics[0] == Topics.ClaimedDelegationReward && len(l.Topics) > 2 && len(l.Data) >= 32 {
			address := common.BytesToAddress(l.Topics[1][12:])
			stakerID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[2][:]).Uint64())
			reward := new(big.Int).SetBytes(l.Data[0:32])

			s.IncDelegationClaimedRewards(DelegationID{address, stakerID}, reward)
			s.IncStakerDelegationsClaimedRewards(stakerID, reward)
		}
		return true
	})
}

func OnNewLog(s *Store, l *types.Log) {
	if l.Address != sfc.ContractAddress {
		return
	}
	// Add new stakers
	if l.Topics[0] == Topics.CreatedValidator && len(l.Topics) > 2 && len(l.Data) >= 32 {
		stakerID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
		address := common.BytesToAddress(l.Topics[2][12:])
		createdEpoch := new(big.Int).SetBytes(l.Data[0:32])
		createdTime := new(big.Int).SetBytes(l.Data[32:64])

		s.SetSfcStaker(stakerID, &SfcStaker{
			CreatedEpoch: idx.Epoch(createdEpoch.Uint64()),
			CreatedTime:  inter.FromUnix(int64(createdTime.Uint64())),
			Address:      address,
		})
	}

	// Add/increase delegations
	if l.Topics[0] == Topics.Delegated && len(l.Topics) > 2 && len(l.Data) >= 32 {
		address := common.BytesToAddress(l.Topics[1][12:])
		toStakerID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[2][:]).Uint64())
		amount := new(big.Int).SetBytes(l.Data[0:32])

		prev := s.GetSfcDelegation(DelegationID{address, toStakerID})
		if prev != nil {
			amount.Add(amount, prev.Amount)
		}
		s.SetSfcDelegation(DelegationID{address, toStakerID}, &SfcDelegation{
			Amount: amount,
		})
	}

	// Deactivate stakes
	if (l.Topics[0] == Topics.DeactivatedValidator) && len(l.Topics) > 1 {
		stakerID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
		deactivatedEpoch := new(big.Int).SetBytes(l.Data[0:32])
		deactivatedTime := new(big.Int).SetBytes(l.Data[32:64])

		staker := s.GetSfcStaker(stakerID)
		if staker == nil {
			return
		}
		staker.DeactivatedEpoch = idx.Epoch(deactivatedEpoch.Uint64())
		staker.DeactivatedTime = inter.FromUnix(int64(deactivatedTime.Uint64()))
		s.SetSfcStaker(stakerID, staker)
	}

	// Change status
	if (l.Topics[0] == Topics.ChangedValidatorStatus) && len(l.Topics) > 1 {
		stakerID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
		status := new(big.Int).SetBytes(l.Data[0:32])

		staker := s.GetSfcStaker(stakerID)
		if staker == nil {
			return
		}
		staker.Status = status.Uint64()
		s.SetSfcStaker(stakerID, staker)
	}

	if l.Topics[0] == Topics.Undelegated && len(l.Topics) > 2 && len(l.Data) >= 32 {
		address := common.BytesToAddress(l.Topics[1][12:])
		toStakerID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[2][:]).Uint64())
		amount := new(big.Int).SetBytes(l.Data[0:32])
		id := DelegationID{address, toStakerID}

		delegation := s.GetSfcDelegation(id)
		if delegation == nil {
			return
		}
		delegation.Amount.Sub(delegation.Amount, amount)
		if delegation.Amount.Sign() > 0 {
			s.SetSfcDelegation(id, delegation)
		} else {
			s.DelSfcDelegation(id)
		}
	}

	// Track rewards
	if (l.Topics[0] == Topics.ClaimedRewards || l.Topics[0] == Topics.RestakedRewards) && len(l.Topics) > 2 && len(l.Data) >= 96 {
		address := common.BytesToAddress(l.Topics[1][12:])
		stakerID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[2][:]).Uint64())
		reward0 := new(big.Int).SetBytes(l.Data[0:32])
		reward1 := new(big.Int).SetBytes(l.Data[32:64])
		reward2 := new(big.Int).SetBytes(l.Data[64:96])
		reward := new(big.Int).Add(reward0.Add(reward0, reward1), reward2)

		s.IncDelegationClaimedRewards(DelegationID{address, stakerID}, reward)
		s.IncStakerDelegationsClaimedRewards(stakerID, reward)
	}
}
