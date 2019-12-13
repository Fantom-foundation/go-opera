package gossip

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/sfctype"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis/sfc"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis/sfc/sfcpos"
	"github.com/Fantom-foundation/go-lachesis/utils"
)

// GetActiveSfcStakers returns stakers which will become validators in next epoch
func (s *Service) GetActiveSfcStakers() []sfctype.SfcStakerAndID {
	stakers := make([]sfctype.SfcStakerAndID, 0, 200)
	s.store.ForEachSfcStaker(func(it sfctype.SfcStakerAndID) {
		if it.Staker.DeactivatedEpoch == 0 && !it.Staker.IsCheater {
			stakers = append(stakers, it)
		}
	})
	return stakers
}

func (s *Service) delAllStakerData(stakerID idx.StakerID) {
	s.store.DelSfcStaker(stakerID)
	s.store.ResetBlocksMissed(stakerID)
	s.store.DelActiveValidationScore(stakerID)
	s.store.DelDirtyValidationScore(stakerID)
	s.store.DelActiveOriginationScore(stakerID)
	s.store.DelDirtyOriginationScore(stakerID)
	s.store.DelWeightedDelegatorsFee(stakerID)
	s.store.DelStakerPOI(stakerID)
	s.store.DelStakerClaimedRewards(stakerID)
	s.store.DelStakerDelegatorsClaimedRewards(stakerID)
}

func (s *Service) delAllDelegatorData(address common.Address) {
	s.store.DelSfcDelegator(address)
	s.store.DelDelegatorClaimedRewards(address)
}

func (s *Service) calcValidatingPowers(stakers []sfctype.SfcStakerAndID) []*big.Int {
	validatingPowers := make([]*big.Int, 0, 200)
	scores := make([]*big.Int, 0, 200)
	pois := make([]*big.Int, 0, 200)
	stakes := make([]*big.Int, 0, 200)

	totalStake := new(big.Int)
	totalPoI := new(big.Int)
	totalScore := new(big.Int)

	for _, it := range stakers {
		stake := it.Staker.CalcTotalStake()
		poi := s.store.GetStakerPOI(it.StakerID)
		if poi.Sign() == 0 {
			poi = big.NewInt(1)
		}
		// score = OriginationScore + ValidationScore
		score := s.store.GetActiveOriginationScore(it.StakerID)
		score.Add(score, s.store.GetActiveValidationScore(it.StakerID))
		if score.Sign() == 0 {
			score = big.NewInt(1)
		}

		stakes = append(stakes, stake)
		scores = append(scores, score)
		pois = append(pois, poi)

		totalStake.Add(totalStake, stake)
		totalPoI.Add(totalPoI, poi)
		totalScore.Add(totalPoI, score)
	}

	for i := range stakers {
		// validatingPower = ((1 - CONST) * stake + (CONST) * PoI) * score,
		// where PoI is rebased to be comparable with stake, score is rebased to [0, 1]
		stakeWithRatio := stakes[i]
		stakeWithRatio.Mul(stakeWithRatio, new(big.Int).Sub(lachesis.PercentUnit, s.config.Net.Economy.ValidatorPoiImpact))
		stakeWithRatio.Div(stakeWithRatio, lachesis.PercentUnit)

		poiRebased := pois[i]
		poiRebased.Mul(poiRebased, totalStake)
		poiRebased.Div(poiRebased, totalPoI)

		poiRebasedWithRatio := poiRebased
		poiRebasedWithRatio.Mul(poiRebasedWithRatio, s.config.Net.Economy.ValidatorPoiImpact)
		poiRebasedWithRatio.Div(poiRebasedWithRatio, lachesis.PercentUnit)

		validatingPower := new(big.Int)
		validatingPower.Add(stakeWithRatio, poiRebasedWithRatio)
		validatingPower.Mul(validatingPower, scores[i])
		validatingPower.Div(validatingPower, totalScore)

		validatingPowers = append(validatingPowers, validatingPower)
	}

	return validatingPowers
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
			if l.Topics[0] == sfcpos.Topics.CreatedStake && len(l.Topics) > 2 && len(l.Data) >= 32 {
				stakerID := idx.StakerID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
				address := common.BytesToAddress(l.Topics[2][12:])
				amount := new(big.Int).SetBytes(l.Data[0:32])

				s.store.SetSfcStaker(stakerID, &sfctype.SfcStaker{
					Address:      address,
					CreatedEpoch: epoch,
					CreatedTime:  block.Time,
					StakeAmount:  amount,
					DelegatedMe:  big.NewInt(0),
				})
			}

			// Increase stakes
			if l.Topics[0] == sfcpos.Topics.IncreasedStake && len(l.Topics) > 1 && len(l.Data) >= 32 {
				stakerID := idx.StakerID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
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
			if l.Topics[0] == sfcpos.Topics.CreatedDelegation && len(l.Topics) > 1 && len(l.Data) >= 32 {
				address := common.BytesToAddress(l.Topics[1][12:])
				toStakerID := idx.StakerID(new(big.Int).SetBytes(l.Topics[2][12:]).Uint64())
				amount := new(big.Int).SetBytes(l.Data[0:32])

				staker := s.store.GetSfcStaker(toStakerID)
				if staker == nil {
					s.Log.Error("Internal SFC index isn't synced with SFC contract")
					continue
				}
				staker.DelegatedMe.Add(staker.DelegatedMe, amount)

				s.store.SetSfcDelegator(address, &sfctype.SfcDelegator{
					ToStakerID:   toStakerID,
					CreatedEpoch: epoch,
					CreatedTime:  block.Time,
					Amount:       amount,
				})
				s.store.SetSfcStaker(toStakerID, staker)
			}

			// Deactivate stakes
			if l.Topics[0] == sfcpos.Topics.PreparedToWithdrawStake && len(l.Topics) > 1 {
				stakerID := idx.StakerID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())

				staker := s.store.GetSfcStaker(stakerID)
				staker.DeactivatedEpoch = epoch
				staker.DeactivatedTime = block.Time
				s.store.SetSfcStaker(stakerID, staker)
			}

			// Deactivate delegators
			if l.Topics[0] == sfcpos.Topics.PreparedToWithdrawDelegation && len(l.Topics) > 1 {
				address := common.BytesToAddress(l.Topics[1][12:])

				delegator := s.store.GetSfcDelegator(address)
				staker := s.store.GetSfcStaker(delegator.ToStakerID)
				if staker != nil {
					staker.DelegatedMe.Sub(staker.DelegatedMe, delegator.Amount)
					s.store.SetSfcStaker(delegator.ToStakerID, staker)
				}
				delegator.DeactivatedEpoch = epoch
				delegator.DeactivatedTime = block.Time
				s.store.SetSfcDelegator(address, delegator)
			}

			// Delete stakes
			if l.Topics[0] == sfcpos.Topics.WithdrawnStake && len(l.Topics) > 1 {
				stakerID := idx.StakerID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
				s.delAllStakerData(stakerID)
			}

			// Delete delegators
			if l.Topics[0] == sfcpos.Topics.WithdrawnDelegation && len(l.Topics) > 1 {
				address := common.BytesToAddress(l.Topics[1][12:])
				s.delAllDelegatorData(address)
			}

			// Track rewards (API-only)
			if l.Topics[0] == sfcpos.Topics.ClaimedValidatorReward && len(l.Topics) > 1 && len(l.Data) >= 32 {
				stakerID := idx.StakerID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
				reward := new(big.Int).SetBytes(l.Data[0:32])

				s.store.IncStakerClaimedRewards(stakerID, reward)
			}
			if l.Topics[0] == sfcpos.Topics.ClaimedDelegationReward && len(l.Topics) > 2 && len(l.Data) >= 32 {
				address := common.BytesToAddress(l.Topics[1][12:])
				stakerID := idx.StakerID(new(big.Int).SetBytes(l.Topics[2][:]).Uint64())
				reward := new(big.Int).SetBytes(l.Data[0:32])

				s.store.IncDelegatorClaimedRewards(address, reward)
				s.store.IncStakerDelegatorsClaimedRewards(stakerID, reward)
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
		s.store.SetDirtyEpochStats(&sfctype.EpochStats{
			Start:    block.Time,
			TotalFee: new(big.Int),
		})
	} else {
		s.store.SetDirtyEpochStats(stats)
	}

	// Write cheaters
	for _, stakerID := range cheaters {
		staker := s.store.GetSfcStaker(stakerID)
		if !staker.IsCheater {
			// write into node DB
			staker.IsCheater = true
			s.store.SetSfcStaker(stakerID, staker)
			// write into SFC contract
			position := sfcpos.Staker(stakerID)
			statedb.SetState(sfc.ContractAddress, position.IsCheater(), common.BytesToHash([]byte{1}))
		}
	}

	if sealEpoch {

		epoch256 := utils.U64to256(uint64(epoch))
		statedb.SetState(sfc.ContractAddress, sfcpos.CurrentSealedEpoch(), epoch256)

		// Write epoch snapshot (for reward)
		cheatersSet := cheaters.Set()
		epochPos := sfcpos.EpochSnapshot(epoch)
		epochValidators := s.store.GetEpochValidators(epoch)
		validatingPowers := s.calcValidatingPowers(epochValidators)

		totalValidatingPower := new(big.Int)
		for i, it := range epochValidators {
			if _, ok := cheatersSet[it.StakerID]; ok {
				continue // don't give reward to cheaters
			}

			validatingPower := validatingPowers[i]

			meritPos := epochPos.ValidatorMerit(it.StakerID)

			statedb.SetState(sfc.ContractAddress, meritPos.StakeAmount(), utils.BigTo256(it.Staker.StakeAmount))
			statedb.SetState(sfc.ContractAddress, meritPos.DelegatedMe(), utils.BigTo256(it.Staker.DelegatedMe))
			statedb.SetState(sfc.ContractAddress, meritPos.ValidatingPower(), utils.BigTo256(validatingPower))

			totalValidatingPower.Add(totalValidatingPower, validatingPower)
		}
		statedb.SetState(sfc.ContractAddress, epochPos.TotalValidatingPower(), utils.BigTo256(totalValidatingPower))
		statedb.SetState(sfc.ContractAddress, epochPos.EpochFee(), utils.BigTo256(stats.TotalFee))
		statedb.SetState(sfc.ContractAddress, epochPos.EndTime(), utils.U64to256(uint64(stats.End.Unix())))
		statedb.SetState(sfc.ContractAddress, epochPos.Duration(), utils.U64to256(uint64((stats.End - stats.Start).Unix())))

		// Add balance for SFC to pay rewards
		blockRewards := big.NewInt(stats.End.Unix() - stats.Start.Unix())
		blockRewards.Mul(blockRewards, s.config.Net.Economy.RewardPerSecond)
		statedb.AddBalance(sfc.ContractAddress, new(big.Int).Add(blockRewards, stats.TotalFee))

		// Select new validators
		for _, it := range s.GetActiveSfcStakers() {
			// Note: cheaters are not active
			if _, ok := cheatersSet[it.StakerID]; ok {
				s.Log.Crit("Cheaters must be erased from Stakers table")
			}
			s.store.SetEpochValidator(epoch+1, it.StakerID, it.Staker)
		}
	}
}
