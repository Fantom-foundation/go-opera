package gossip

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
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
	s.app.ForEachSfcStaker(func(it sfctype.SfcStakerAndID) {
		if it.Staker.Ok() {
			stakers = append(stakers, it)
		}
	})
	return stakers
}

func (s *Service) delAllStakerData(stakerID idx.StakerID) {
	s.app.DelSfcStaker(stakerID)
	s.app.ResetBlocksMissed(stakerID)
	s.app.DelActiveValidationScore(stakerID)
	s.app.DelDirtyValidationScore(stakerID)
	s.app.DelActiveOriginationScore(stakerID)
	s.app.DelDirtyOriginationScore(stakerID)
	s.app.DelWeightedDelegatorsFee(stakerID)
	s.app.DelStakerPOI(stakerID)
	s.app.DelStakerClaimedRewards(stakerID)
	s.app.DelStakerDelegatorsClaimedRewards(stakerID)
}

func (s *Service) delAllDelegatorData(address common.Address) {
	s.app.DelSfcDelegator(address)
	s.app.DelDelegatorClaimedRewards(address)
}

var (
	max128 = new(big.Int).Sub(math.BigPow(2, 128), common.Big1)
)

func (s *Service) calcRewardWeights(stakers []sfctype.SfcStakerAndID, _epochDuration inter.Timestamp) (baseRewardWeights []*big.Int, txRewardWeights []*big.Int) {
	validationScores := make([]*big.Int, 0, len(stakers))
	originationScores := make([]*big.Int, 0, len(stakers))
	pois := make([]*big.Int, 0, len(stakers))
	stakes := make([]*big.Int, 0, len(stakers))

	if _epochDuration == 0 {
		_epochDuration = 1
	}
	epochDuration := new(big.Int).SetUint64(uint64(_epochDuration))

	for _, it := range stakers {
		stake := it.Staker.CalcTotalStake()
		poi := s.app.GetStakerPOI(it.StakerID)
		validationScore := s.app.GetActiveValidationScore(it.StakerID)
		originationScore := s.app.GetActiveOriginationScore(it.StakerID)

		stakes = append(stakes, stake)
		validationScores = append(validationScores, validationScore)
		originationScores = append(originationScores, originationScore)
		pois = append(pois, poi)
	}

	txRewardWeights = make([]*big.Int, 0, len(stakers))
	for i := range stakers {
		// txRewardWeight = ({origination score} + {CONST} * {PoI}) * {validation score}
		// origination score is roughly proportional to {validation score} * {stake}, so the whole formula is roughly
		// {stake} * {validation score} ^ 2
		poiWithRatio := new(big.Int).Mul(pois[i], s.config.Net.Economy.TxRewardPoiImpact)
		poiWithRatio.Div(poiWithRatio, lachesis.PercentUnit)

		txRewardWeight := new(big.Int).Add(originationScores[i], poiWithRatio)
		txRewardWeight.Mul(txRewardWeight, validationScores[i])
		txRewardWeight.Div(txRewardWeight, epochDuration)
		if txRewardWeight.Cmp(max128) > 0 {
			txRewardWeight = new(big.Int).Set(max128) // never going to get here
		}

		txRewardWeights = append(txRewardWeights, txRewardWeight)
	}

	baseRewardWeights = make([]*big.Int, 0, len(stakers))
	for i := range stakers {
		// baseRewardWeight = {stake} * {validationScore ^ 2}
		baseRewardWeight := new(big.Int).Set(stakes[i])
		for pow := 0; pow < 2; pow++ {
			baseRewardWeight.Mul(baseRewardWeight, validationScores[i])
			baseRewardWeight.Div(baseRewardWeight, epochDuration)
		}
		if baseRewardWeight.Cmp(max128) > 0 {
			baseRewardWeight = new(big.Int).Set(max128) // never going to get here
		}

		baseRewardWeights = append(baseRewardWeights, baseRewardWeight)
	}

	return baseRewardWeights, txRewardWeights
}

// getRewardPerSec returns current rewardPerSec, depending on config and value provided by SFC
func (s *Service) getRewardPerSec() *big.Int {
	rewardPerSecond := s.app.GetSfcConstants(s.engine.GetEpoch() - 1).BaseRewardPerSec
	if rewardPerSecond == nil || rewardPerSecond.Sign() == 0 {
		rewardPerSecond = s.config.Net.Economy.InitialRewardPerSecond
	}
	if rewardPerSecond.Cmp(s.config.Net.Economy.MaxRewardPerSecond) > 0 {
		rewardPerSecond = s.config.Net.Economy.MaxRewardPerSecond
	}
	return new(big.Int).Set(rewardPerSecond)
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

				s.app.SetSfcStaker(stakerID, &sfctype.SfcStaker{
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

				staker := s.app.GetSfcStaker(stakerID)
				if staker == nil {
					s.Log.Warn("Internal SFC index isn't synced with SFC contract")
					continue
				}
				staker.StakeAmount = newAmount
				s.app.SetSfcStaker(stakerID, staker)
			}

			// Add new delegators
			if l.Topics[0] == sfcpos.Topics.CreatedDelegation && len(l.Topics) > 1 && len(l.Data) >= 32 {
				address := common.BytesToAddress(l.Topics[1][12:])
				toStakerID := idx.StakerID(new(big.Int).SetBytes(l.Topics[2][:]).Uint64())
				amount := new(big.Int).SetBytes(l.Data[0:32])

				staker := s.app.GetSfcStaker(toStakerID)
				if staker == nil {
					s.Log.Warn("Internal SFC index isn't synced with SFC contract")
					continue
				}
				staker.DelegatedMe.Add(staker.DelegatedMe, amount)

				s.app.SetSfcDelegator(address, &sfctype.SfcDelegator{
					ToStakerID:   toStakerID,
					CreatedEpoch: epoch,
					CreatedTime:  block.Time,
					Amount:       amount,
				})
				s.app.SetSfcStaker(toStakerID, staker)
			}

			// Deactivate stakes
			if (l.Topics[0] == sfcpos.Topics.DeactivatedStake || l.Topics[0] == sfcpos.Topics.PreparedToWithdrawStake) && len(l.Topics) > 1 {
				stakerID := idx.StakerID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())

				staker := s.app.GetSfcStaker(stakerID)
				staker.DeactivatedEpoch = epoch
				staker.DeactivatedTime = block.Time
				s.app.SetSfcStaker(stakerID, staker)
			}

			// Deactivate delegators
			if (l.Topics[0] == sfcpos.Topics.DeactivatedDelegation || l.Topics[0] == sfcpos.Topics.PreparedToWithdrawDelegation) && len(l.Topics) > 1 {
				address := common.BytesToAddress(l.Topics[1][12:])

				delegator := s.app.GetSfcDelegator(address)
				staker := s.app.GetSfcStaker(delegator.ToStakerID)
				if staker != nil {
					staker.DelegatedMe.Sub(staker.DelegatedMe, delegator.Amount)
					if staker.DelegatedMe.Sign() < 0 {
						staker.DelegatedMe = big.NewInt(0)
					}
					s.app.SetSfcStaker(delegator.ToStakerID, staker)
				}
				delegator.DeactivatedEpoch = epoch
				delegator.DeactivatedTime = block.Time
				s.app.SetSfcDelegator(address, delegator)
			}

			// Update stake
			if l.Topics[0] == sfcpos.Topics.UpdatedStake && len(l.Topics) > 1 && len(l.Data) >= 64 {
				stakerID := idx.StakerID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
				newAmount := new(big.Int).SetBytes(l.Data[0:32])
				newDelegatedMe := new(big.Int).SetBytes(l.Data[32:64])

				staker := s.app.GetSfcStaker(stakerID)
				if staker == nil {
					s.Log.Warn("Internal SFC index isn't synced with SFC contract")
					continue
				}
				staker.StakeAmount = newAmount
				staker.DelegatedMe = newDelegatedMe
				s.app.SetSfcStaker(stakerID, staker)
			}

			// Update delegation
			if l.Topics[0] == sfcpos.Topics.UpdatedDelegation && len(l.Topics) > 3 && len(l.Data) >= 32 {
				address := common.BytesToAddress(l.Topics[1][12:])
				newStakerID := idx.StakerID(new(big.Int).SetBytes(l.Topics[3][:]).Uint64())
				newAmount := new(big.Int).SetBytes(l.Data[0:32])

				delegator := s.app.GetSfcDelegator(address)
				if delegator == nil {
					s.Log.Warn("Internal SFC index isn't synced with SFC contract")
					continue
				}
				delegator.Amount = newAmount
				delegator.ToStakerID = newStakerID
				s.app.SetSfcDelegator(address, delegator)
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

			// Track changes of constants by SFC
			if l.Topics[0] == sfcpos.Topics.UpdatedBaseRewardPerSec && len(l.Data) >= 32 {
				baseRewardPerSec := new(big.Int).SetBytes(l.Data[0:32])
				constants := s.app.GetSfcConstants(epoch)
				constants.BaseRewardPerSec = baseRewardPerSec
				s.app.SetSfcConstants(epoch, constants)
			}
			if l.Topics[0] == sfcpos.Topics.UpdatedGasPowerAllocationRate && len(l.Data) >= 64 {
				shortAllocationRate := new(big.Int).SetBytes(l.Data[0:32])
				longAllocationRate := new(big.Int).SetBytes(l.Data[32:64])
				constants := s.app.GetSfcConstants(epoch)
				constants.ShortGasPowerAllocPerSec = shortAllocationRate.Uint64()
				constants.LongGasPowerAllocPerSec = longAllocationRate.Uint64()
				s.app.SetSfcConstants(epoch, constants)
			}

			// Track rewards (API-only)
			if l.Topics[0] == sfcpos.Topics.ClaimedValidatorReward && len(l.Topics) > 1 && len(l.Data) >= 32 {
				stakerID := idx.StakerID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
				reward := new(big.Int).SetBytes(l.Data[0:32])

				s.app.IncStakerClaimedRewards(stakerID, reward)
			}
			if l.Topics[0] == sfcpos.Topics.ClaimedDelegationReward && len(l.Topics) > 2 && len(l.Data) >= 32 {
				address := common.BytesToAddress(l.Topics[1][12:])
				stakerID := idx.StakerID(new(big.Int).SetBytes(l.Topics[2][:]).Uint64())
				reward := new(big.Int).SetBytes(l.Data[0:32])

				s.app.IncDelegatorClaimedRewards(address, reward)
				s.app.IncStakerDelegatorsClaimedRewards(stakerID, reward)
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
		staker := s.app.GetSfcStaker(stakerID)
		if staker.HasFork() {
			continue
		}
		// write into DB
		staker.Status |= sfctype.ForkBit
		s.app.SetSfcStaker(stakerID, staker)
		// write into SFC contract
		position := sfcpos.Staker(stakerID)
		statedb.SetState(sfc.ContractAddress, position.Status(), utils.U64to256(staker.Status))
	}

	if sealEpoch {
		if s.app.HasSfcConstants(epoch) {
			s.app.SetSfcConstants(epoch+1, s.app.GetSfcConstants(epoch))
		}

		// Write offline validators
		for _, it := range s.app.GetSfcStakers() {
			if it.Staker.Offline() {
				continue
			}

			gotMissed := s.app.GetBlocksMissed(it.StakerID)
			badMissed := s.config.Net.Economy.OfflinePenaltyThreshold
			if gotMissed.Num >= badMissed.BlocksNum && gotMissed.Period >= inter.Timestamp(badMissed.Period) {
				// write into DB
				it.Staker.Status |= sfctype.OfflineBit
				s.app.SetSfcStaker(it.StakerID, it.Staker)
				// write into SFC contract
				position := sfcpos.Staker(it.StakerID)
				statedb.SetState(sfc.ContractAddress, position.Status(), utils.U64to256(it.Staker.Status))
			}
		}

		// Write epoch snapshot (for reward)
		cheatersSet := cheaters.Set()
		epochPos := sfcpos.EpochSnapshot(epoch)
		epochValidators := s.app.GetEpochValidators(epoch)
		baseRewardWeights, txRewardWeights := s.calcRewardWeights(epochValidators, stats.Duration())

		totalBaseRewardWeight := new(big.Int)
		totalTxRewardWeight := new(big.Int)
		totalStake := new(big.Int)
		totalDelegated := new(big.Int)
		for i, it := range epochValidators {
			baseRewardWeight := baseRewardWeights[i]
			txRewardWeight := txRewardWeights[i]
			totalStake.Add(totalStake, it.Staker.StakeAmount)
			totalDelegated.Add(totalDelegated, it.Staker.DelegatedMe)

			if _, ok := cheatersSet[it.StakerID]; ok {
				continue // don't give reward to cheaters
			}
			if baseRewardWeight.Sign() == 0 && txRewardWeight.Sign() == 0 {
				continue // don't give reward to offline validators
			}

			meritPos := epochPos.ValidatorMerit(it.StakerID)

			statedb.SetState(sfc.ContractAddress, meritPos.StakeAmount(), utils.BigTo256(it.Staker.StakeAmount))
			statedb.SetState(sfc.ContractAddress, meritPos.DelegatedMe(), utils.BigTo256(it.Staker.DelegatedMe))
			statedb.SetState(sfc.ContractAddress, meritPos.BaseRewardWeight(), utils.BigTo256(baseRewardWeight))
			statedb.SetState(sfc.ContractAddress, meritPos.TxRewardWeight(), utils.BigTo256(txRewardWeight))

			totalBaseRewardWeight.Add(totalBaseRewardWeight, baseRewardWeight)
			totalTxRewardWeight.Add(totalTxRewardWeight, txRewardWeight)
		}
		baseRewardPerSec := s.getRewardPerSec()

		// set total supply
		baseRewards := new(big.Int).Mul(big.NewInt(stats.Duration().Unix()), baseRewardPerSec)
		rewards := new(big.Int).Add(baseRewards, stats.TotalFee)
		totalSupply := new(big.Int).Add(s.app.GetTotalSupply(), rewards)
		statedb.SetState(sfc.ContractAddress, sfcpos.CurrentSealedEpoch(), utils.U64to256(uint64(epoch)))
		s.app.SetTotalSupply(totalSupply)

		statedb.SetState(sfc.ContractAddress, epochPos.TotalBaseRewardWeight(), utils.BigTo256(totalBaseRewardWeight))
		statedb.SetState(sfc.ContractAddress, epochPos.TotalTxRewardWeight(), utils.BigTo256(totalTxRewardWeight))
		statedb.SetState(sfc.ContractAddress, epochPos.EpochFee(), utils.BigTo256(stats.TotalFee))
		statedb.SetState(sfc.ContractAddress, epochPos.EndTime(), utils.U64to256(uint64(stats.End.Unix())))
		statedb.SetState(sfc.ContractAddress, epochPos.Duration(), utils.U64to256(uint64(stats.Duration().Unix())))
		statedb.SetState(sfc.ContractAddress, epochPos.BaseRewardPerSecond(), utils.BigTo256(baseRewardPerSec))
		statedb.SetState(sfc.ContractAddress, epochPos.StakeTotalAmount(), utils.BigTo256(totalStake))
		statedb.SetState(sfc.ContractAddress, epochPos.DelegationsTotalAmount(), utils.BigTo256(totalDelegated))
		statedb.SetState(sfc.ContractAddress, epochPos.TotalSupply(), utils.BigTo256(totalSupply))
		statedb.SetState(sfc.ContractAddress, sfcpos.CurrentSealedEpoch(), utils.U64to256(uint64(epoch)))

		// Add balance for SFC to pay rewards
		statedb.AddBalance(sfc.ContractAddress, rewards)

		// Select new validators
		s.app.SetEpochValidators(epoch+1, s.GetActiveSfcStakers())
	}
}
