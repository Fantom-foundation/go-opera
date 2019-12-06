package gossip

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis/sfc"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis/sfc/sfcpos"
	"github.com/Fantom-foundation/go-lachesis/utils"
)

// SfcStaker is the node-side representation of SFC staker
type SfcStaker struct {
	CreatedEpoch idx.Epoch
	CreatedTime  inter.Timestamp

	StakeAmount *big.Int
	DelegatedMe *big.Int

	Address common.Address
}

// SfcStakerAndID is pair SfcStaker + StakerID
type SfcStakerAndID struct {
	StakerID uint64
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

	Amount *big.Int

	ToStakerID uint64
}

// processSfc applies the new SFC state
func (s *Service) processSfc(block *inter.Block, receipts types.Receipts, blockFee *big.Int, sealEpoch bool, cheaters inter.Cheaters, statedb *state.StateDB) {
	// s.engineMu is locked here

	// process SFC contract logs
	epoch := s.engine.GetEpoch()
	for _, receipt := range receipts {
		for _, l := range receipt.Logs {
			if l.Address != sfc.ContractAddress {
				continue
			}
			// Add new stakers
			if l.Topics[0] == sfcpos.CreateStakeTopic {
				stakerID := new(big.Int).SetBytes(l.Topics[1][:]).Uint64()
				address := common.BytesToAddress(l.Topics[2][12:])
				amount := new(big.Int).SetBytes(l.Data[0:32])

				s.store.SetSfcStaker(stakerID, &SfcStaker{
					Address:      address,
					CreatedEpoch: epoch,
					CreatedTime:  block.Time,
					StakeAmount:  amount,
					DelegatedMe:  big.NewInt(0),
				})
			}

			// Increase stakes
			if l.Topics[0] == sfcpos.IncreasedStakeTopic {
				stakerID := new(big.Int).SetBytes(l.Topics[1][:]).Uint64()
				newAmount := new(big.Int).SetBytes(l.Data[0:32])

				staker := s.store.GetSfcStaker(stakerID)
				if staker == nil {
					s.Log.Error("Internal SFC index isn't synced with SFC contract")
					continue
				}
				staker.StakeAmount = newAmount
				s.store.SetSfcStaker(stakerID, staker)
			}

			// Add new delegators
			if l.Topics[0] == sfcpos.CreatedDelegationTopic {
				address := common.BytesToAddress(l.Topics[1][12:])
				toStakerID := new(big.Int).SetBytes(l.Topics[2][12:]).Uint64()
				amount := new(big.Int).SetBytes(l.Data[0:32])

				staker := s.store.GetSfcStaker(toStakerID)
				if staker == nil {
					s.Log.Error("Internal SFC index isn't synced with SFC contract")
					continue
				}
				staker.DelegatedMe.Add(staker.DelegatedMe, amount)

				s.store.SetSfcDelegator(address, &SfcDelegator{
					ToStakerID:   toStakerID,
					CreatedEpoch: epoch,
					CreatedTime:  block.Time,
					Amount:       amount,
				})
				s.store.SetSfcStaker(toStakerID, staker)
			}

			// Delete stakes
			if l.Topics[0] == sfcpos.DeactivateStakeTopic {
				stakerID := new(big.Int).SetBytes(l.Topics[1][:]).Uint64()
				s.store.DelSfcStaker(stakerID)
			}

			// Delete delegators
			if l.Topics[0] == sfcpos.DeactivateDelegationTopic {
				address := common.BytesToAddress(l.Topics[1][12:])

				delegator := s.store.GetSfcDelegator(address)
				staker := s.store.GetSfcStaker(delegator.ToStakerID)
				if staker != nil {
					staker.DelegatedMe.Sub(staker.DelegatedMe, delegator.Amount)
					s.store.SetSfcStaker(delegator.ToStakerID, staker)
				}
				s.store.DelSfcDelegator(address)
			}
		}
	}

	// Update EpochStats
	stats := s.store.GetDirtyEpochStats()
	stats.TotalFee = new(big.Int).Add(stats.TotalFee, blockFee)
	if sealEpoch {
		// dirty EpochStats becomes active
		stats.End = block.Time
		s.store.SetEpochStats(epoch, stats)

		// new dirty EpochStats
		s.store.SetDirtyEpochStats(&EpochStats{
			Start:    block.Time,
			TotalFee: new(big.Int),
		})
	} else {
		s.store.SetDirtyEpochStats(stats)
	}

	// Write cheaters into SFC
	/*for _, validator := range cheaters {
		position := sfcpos.VStake(validator)
		if statedb.GetState(sfc.ContractAddress, position.IsCheater()) == (common.Hash{}) {
			statedb.SetState(sfc.ContractAddress, position.IsCheater(), common.BytesToHash([]byte{1}))
		}
	}*/

	if sealEpoch {

		epoch256 := utils.U64to256(uint64(epoch))
		statedb.SetState(sfc.ContractAddress, sfcpos.CurrentSealedEpoch(), epoch256)

		// Write epoch snapshot (for reward)
		cheatersSet := cheaters.Set()
		epochPos := sfcpos.EpochSnapshot(epoch)
		totalValidatingPower := new(big.Int)
		for _, it := range s.store.GetEpochValidators(epoch) {
			if _, ok := cheatersSet[it.Staker.Address]; ok {
				continue // don't give reward to cheaters
			}

			meritPos := epochPos.ValidatorMerit(it.Staker.Address)

			validatingPower := it.Staker.CalcTotalStake() // TODO

			statedb.SetState(sfc.ContractAddress, meritPos.StakeAmount(), utils.BigTo256(it.Staker.StakeAmount))
			statedb.SetState(sfc.ContractAddress, meritPos.DelegatedMe(), utils.BigTo256(it.Staker.DelegatedMe))
			statedb.SetState(sfc.ContractAddress, meritPos.ValidatingPower(), utils.BigTo256(validatingPower))

			totalValidatingPower.Add(totalValidatingPower, validatingPower)
		}
		statedb.SetState(sfc.ContractAddress, epochPos.TotalValidatingPower(), utils.BigTo256(totalValidatingPower))
		statedb.SetState(sfc.ContractAddress, epochPos.EpochFee(), utils.BigTo256(stats.TotalFee))
		statedb.SetState(sfc.ContractAddress, epochPos.EndTime(), utils.U64to256(uint64(stats.End.Unix())))
		statedb.SetState(sfc.ContractAddress, epochPos.Duration(), utils.U64to256(uint64((stats.End - stats.Start).Unix())))

		// Erase cheaters from our stakers index
		/*for _, validator := range cheaters {
			s.store.DelSfcStaker(validator)
		}*/
		// Select new validators
		for _, it := range s.store.GetSfcStakers() {
			// Note: cheaters are already erased from Stakers table
			if _, ok := cheatersSet[it.Staker.Address]; ok {
				s.Log.Crit("Cheaters must be erased from Stakers table")
			}
			s.store.SetEpochValidator(epoch+1, it.StakerID, it.Staker)
		}
	}
}
