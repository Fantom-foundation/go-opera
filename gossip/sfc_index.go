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

// SfcConstants are constants which may be changed by SFC contract
type SfcConstants struct {
	ShortGasPowerAllocPerSec uint64
	LongGasPowerAllocPerSec  uint64
	BaseRewardPerSec         *big.Int
}

// GetActiveSfcStakers returns stakers which will become validators in next epoch
func (s *Service) GetActiveSfcStakers() []sfctype.SfcStakerAndID {
	stakers := make([]sfctype.SfcStakerAndID, 0, 200)
	s.store.ForEachSfcStaker(func(it sfctype.SfcStakerAndID) {
		if it.Staker.Ok() {
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
		poi := s.store.GetStakerPOI(it.StakerID)
		validationScore := s.store.GetActiveValidationScore(it.StakerID)
		originationScore := s.store.GetActiveOriginationScore(it.StakerID)

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
	rewardPerSecond := s.store.GetSfcConstants(s.engine.GetEpoch() - 1).BaseRewardPerSec
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

			// Track changes of constants by SFC
			if l.Topics[0] == sfcpos.Topics.UpdatedBaseRewardPerSec && len(l.Data) >= 32 {
				baseRewardPerSec := new(big.Int).SetBytes(l.Data[0:32])
				constants := s.store.GetSfcConstants(epoch)
				constants.BaseRewardPerSec = baseRewardPerSec
				s.store.SetSfcConstants(epoch, constants)
			}
			if l.Topics[0] == sfcpos.Topics.UpdatedGasPowerAllocationRate && len(l.Data) >= 64 {
				shortAllocationRate := new(big.Int).SetBytes(l.Data[0:32])
				longAllocationRate := new(big.Int).SetBytes(l.Data[32:64])
				constants := s.store.GetSfcConstants(epoch)
				constants.ShortGasPowerAllocPerSec = shortAllocationRate.Uint64()
				constants.LongGasPowerAllocPerSec = longAllocationRate.Uint64()
				s.store.SetSfcConstants(epoch, constants)
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
		if staker.HasFork() {
			continue
		}
		// write into DB
		staker.Status |= sfctype.ForkBit
		s.store.SetSfcStaker(stakerID, staker)
		// write into SFC contract
		position := sfcpos.Staker(stakerID)
		statedb.SetState(sfc.ContractAddress, position.Status(), utils.U64to256(staker.Status))
	}

	if sealEpoch {
		epoch256 := utils.U64to256(uint64(epoch))
		statedb.SetState(sfc.ContractAddress, sfcpos.CurrentSealedEpoch(), epoch256)

		// Write offline validators
		for _, it := range s.store.GetSfcStakers() {
			if it.Staker.Offline() {
				continue
			}

			gotMissed := s.store.GetBlocksMissed(it.StakerID)
			badMissed := s.config.Net.Economy.OfflinePenaltyThreshold
			if gotMissed.Num >= badMissed.Num && gotMissed.Period >= inter.Timestamp(badMissed.Period) {
				// write into DB
				it.Staker.Status |= sfctype.OfflineBit
				s.store.SetSfcStaker(it.StakerID, it.Staker)
				// write into SFC contract
				position := sfcpos.Staker(it.StakerID)
				statedb.SetState(sfc.ContractAddress, position.Status(), utils.U64to256(it.Staker.Status))
			}
		}

		// Write epoch snapshot (for reward)
		cheatersSet := cheaters.Set()
		epochPos := sfcpos.EpochSnapshot(epoch)
		epochValidators := s.store.GetEpochValidators(epoch)
		baseRewardWeights, txRewardWeights := s.calcRewardWeights(epochValidators, stats.Duration())

		totalBaseRewardWeight := new(big.Int)
		totalTxRewardWeight := new(big.Int)
		for i, it := range epochValidators {
			baseRewardWeight := baseRewardWeights[i]
			txRewardWeight := txRewardWeights[i]

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

		statedb.SetState(sfc.ContractAddress, epochPos.TotalBaseRewardWeight(), utils.BigTo256(totalBaseRewardWeight))
		statedb.SetState(sfc.ContractAddress, epochPos.TotalTxRewardWeight(), utils.BigTo256(totalTxRewardWeight))
		statedb.SetState(sfc.ContractAddress, epochPos.EpochFee(), utils.BigTo256(stats.TotalFee))
		statedb.SetState(sfc.ContractAddress, epochPos.EndTime(), utils.U64to256(uint64(stats.End.Unix())))
		statedb.SetState(sfc.ContractAddress, epochPos.Duration(), utils.U64to256(uint64(stats.Duration().Unix())))
		statedb.SetState(sfc.ContractAddress, epochPos.BaseRewardPerSecond(), utils.BigTo256(baseRewardPerSec))

		// Add balance for SFC to pay rewards
		baseRewards := new(big.Int).Mul(big.NewInt(stats.Duration().Unix()), baseRewardPerSec)
		statedb.AddBalance(sfc.ContractAddress, new(big.Int).Add(baseRewards, stats.TotalFee))

		// Select new validators
		for _, it := range s.GetActiveSfcStakers() {
			// Note: cheaters are not active
			if _, ok := cheatersSet[it.StakerID]; ok {
				s.Log.Crit("Cheaters must be deactivated")
			}
			s.store.SetEpochValidator(epoch+1, it.StakerID, it.Staker)
		}
	}
}
