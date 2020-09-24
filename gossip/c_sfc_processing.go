package gossip

import (
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/app"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/sfctype"
	"github.com/Fantom-foundation/go-opera/opera/genesis/sfc"
	"github.com/Fantom-foundation/go-opera/opera/genesis/sfc/sfcpos"
	"github.com/Fantom-foundation/go-opera/utils"
)

// GetActiveSfcStakers returns stakers which will become validators in next epoch
func (s *Service) GetActiveSfcStakers() []sfctype.SfcStakerAndID {
	stakers := make([]sfctype.SfcStakerAndID, 0, 200)
	s.store.app.ForEachSfcStaker(func(it sfctype.SfcStakerAndID) {
		if it.Staker.Ok() {
			stakers = append(stakers, it)
		}
	})
	return stakers
}

func (s *Service) delAllStakerData(validatorID idx.ValidatorID) {
	s.store.app.DelSfcStaker(validatorID)
}

var (
	max128 = new(big.Int).Sub(math.BigPow(2, 128), common.Big1)
)

func (s *Service) calcRewardWeights(bs *BlockState, es *EpochState) (baseRewardWeights []*big.Int, txRewardWeights []*big.Int) {
	uptimes := make([]*big.Int, 0, len(bs.ValidatorStates))
	originationScores := make([]*big.Int, 0, len(bs.ValidatorStates))
	stakes := make([]*big.Int, 0, len(bs.ValidatorStates))

	_epochDuration := es.Duration()
	if _epochDuration == 0 {
		_epochDuration = 1
	}
	epochDuration := new(big.Int).SetUint64(uint64(_epochDuration))

	for i, info := range bs.ValidatorStates {
		stake := es.Validators.GetWeightByIdx(idx.Validator(i))

		stakes = append(stakes, new(big.Int).SetUint64(uint64(stake)))
		uptimes = append(uptimes, new(big.Int).SetUint64(uint64(info.Uptime)))
		originationScores = append(originationScores, info.Originated)
	}

	txRewardWeights = make([]*big.Int, 0, len(bs.ValidatorStates))
	for i := range bs.ValidatorStates {
		// txRewardWeight = {origination score} * {uptime score}
		// origination score is roughly proportional to {uptime score} * {stake}, so the whole formula is roughly
		// {stake} * {uptime score} ^ 2
		txRewardWeight := new(big.Int).Mul(originationScores[i], uptimes[i])
		txRewardWeight.Div(txRewardWeight, epochDuration)
		if txRewardWeight.Cmp(max128) > 0 {
			txRewardWeight = new(big.Int).Set(max128) // never going to get here
		}

		txRewardWeights = append(txRewardWeights, txRewardWeight)
	}

	baseRewardWeights = make([]*big.Int, 0, len(bs.ValidatorStates))
	for i := range bs.ValidatorStates {
		// baseRewardWeight = {stake} * {uptime ^ 2}
		baseRewardWeight := new(big.Int).Set(stakes[i])
		for pow := 0; pow < 2; pow++ {
			baseRewardWeight.Mul(baseRewardWeight, uptimes[i])
			baseRewardWeight.Div(baseRewardWeight, epochDuration)
		}
		if baseRewardWeight.Cmp(max128) > 0 {
			baseRewardWeight = new(big.Int).Set(max128) // never going to get here
		}

		baseRewardWeights = append(baseRewardWeights, baseRewardWeight)
	}

	return baseRewardWeights, txRewardWeights
}

func (s *Service) getRewardPerSec() *big.Int {
	return s.config.Net.Economy.InitialRewardPerSecond
}

func (s *Service) getOfflinePenaltyThreshold() app.BlocksMissed {
	return app.BlocksMissed{
		Num:    s.config.Net.Economy.InitialOfflinePenaltyThreshold.BlocksNum,
		Period: inter.Timestamp(s.config.Net.Economy.InitialOfflinePenaltyThreshold.Period),
	}
}

// processSfc applies the new SFC state
func (s *Service) processSfc(bs *BlockState, es *EpochState, block *inter.Block, receipts types.Receipts, blockFee *big.Int, sealEpoch bool, cheaters lachesis.Cheaters, statedb *state.StateDB) {
	// s.engineMu is locked here

	// process SFC contract logs
	epoch := s.store.GetEpoch()
	for _, receipt := range receipts {
		for _, l := range receipt.Logs {
			if l.Address != sfc.ContractAddress {
				continue
			}
			// Add new stakers
			if l.Topics[0] == sfcpos.Topics.CreatedStake && len(l.Topics) > 2 && len(l.Data) >= 32 {
				validatorID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
				address := common.BytesToAddress(l.Topics[2][12:])
				amount := new(big.Int).SetBytes(l.Data[0:32])

				s.store.app.SetSfcStaker(validatorID, &sfctype.SfcStaker{
					Address:      address,
					CreatedEpoch: epoch,
					CreationTime: block.Time,
					StakeAmount:  amount,
					DelegatedMe:  big.NewInt(0),
				})
			}

			// Increase stakes
			if l.Topics[0] == sfcpos.Topics.IncreasedStake && len(l.Topics) > 1 && len(l.Data) >= 32 {
				validatorID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
				newAmount := new(big.Int).SetBytes(l.Data[0:32])

				staker := s.store.app.GetSfcStaker(validatorID)
				if staker == nil {
					s.Log.Warn("Internal SFC index isn't synced with SFC contract")
					continue
				}
				staker.StakeAmount = newAmount
				s.store.app.SetSfcStaker(validatorID, staker)
			}

			// Deactivate stakes
			if (l.Topics[0] == sfcpos.Topics.DeactivatedStake) && len(l.Topics) > 1 {
				validatorID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())

				staker := s.store.app.GetSfcStaker(validatorID)
				staker.DeactivatedEpoch = epoch
				staker.DeactivatedTime = block.Time
				s.store.app.SetSfcStaker(validatorID, staker)
			}

			// Update stake
			if l.Topics[0] == sfcpos.Topics.UpdatedStake && len(l.Topics) > 1 && len(l.Data) >= 64 {
				validatorID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
				newAmount := new(big.Int).SetBytes(l.Data[0:32])
				newDelegatedMe := new(big.Int).SetBytes(l.Data[32:64])

				staker := s.store.app.GetSfcStaker(validatorID)
				if staker == nil {
					s.Log.Warn("Internal SFC index isn't synced with SFC contract")
					continue
				}
				staker.StakeAmount = newAmount
				staker.DelegatedMe = newDelegatedMe
				s.store.app.SetSfcStaker(validatorID, staker)
			}

			// Delete stakes
			if l.Topics[0] == sfcpos.Topics.WithdrawnStake && len(l.Topics) > 1 {
				validatorID := idx.ValidatorID(new(big.Int).SetBytes(l.Topics[1][:]).Uint64())
				s.delAllStakerData(validatorID)
			}
		}
	}

	// Update EpochStats
	epochFee := new(big.Int).Add(bs.EpochFee, blockFee)

	// Write cheaters
	for _, validatorID := range cheaters {
		staker := s.store.app.GetSfcStaker(validatorID)
		if staker.HasFork() {
			continue
		}
		// write into DB
		staker.Status |= sfctype.ForkBit
		s.store.app.SetSfcStaker(validatorID, staker)
		// write into SFC contract
		position := sfcpos.Staker(validatorID)
		statedb.SetState(sfc.ContractAddress, position.Status(), utils.U64to256(staker.Status))
	}

	if sealEpoch {
		// app final uptime for validators
		for _, info := range bs.ValidatorStates {
			if bs.Block-info.PrevBlock <= s.config.Net.Economy.BlockMissedLatency {
				info.Uptime += inter.MaxTimestamp(block.Time, info.PrevMedianTime) - info.PrevMedianTime
			}
		}

		// Write offline validators
		for _, it := range s.store.app.GetSfcStakers() {
			if it.Staker.Offline() {
				continue
			}

			info := bs.GetValidatorState(it.ValidatorID, es.Validators)
			gotMissed := app.BlocksMissed{
				Num:    bs.Block - info.PrevBlock,
				Period: inter.MaxTimestamp(block.Time, info.PrevMedianTime) - info.PrevMedianTime,
			}
			badMissed := s.getOfflinePenaltyThreshold()
			if gotMissed.Num >= badMissed.Num && gotMissed.Period >= badMissed.Period {
				// write into DB
				it.Staker.Status |= sfctype.OfflineBit
				s.store.app.SetSfcStaker(it.ValidatorID, it.Staker)
				// write into SFC contract
				position := sfcpos.Staker(it.ValidatorID)
				statedb.SetState(sfc.ContractAddress, position.Status(), utils.U64to256(it.Staker.Status))
			}
		}

		// Write epoch snapshot (for reward)
		cheatersSet := cheaters.Set()
		epochPos := sfcpos.EpochSnapshot(epoch)
		baseRewardWeights, txRewardWeights := s.calcRewardWeights(bs, es)
		stakers := es.ValidatorProfiles

		totalBaseRewardWeight := new(big.Int)
		totalTxRewardWeight := new(big.Int)
		totalStake := new(big.Int)
		totalDelegated := new(big.Int)
		for i, it := range stakers {
			baseRewardWeight := baseRewardWeights[i]
			txRewardWeight := txRewardWeights[i]
			totalStake.Add(totalStake, it.Staker.StakeAmount)
			totalDelegated.Add(totalDelegated, it.Staker.DelegatedMe)

			if _, ok := cheatersSet[it.ValidatorID]; ok {
				continue // don't give reward to cheaters
			}
			if baseRewardWeight.Sign() == 0 && txRewardWeight.Sign() == 0 {
				continue // don't give reward to offline validators
			}

			meritPos := epochPos.ValidatorMerit(it.ValidatorID)

			statedb.SetState(sfc.ContractAddress, meritPos.StakeAmount(), utils.BigTo256(it.Staker.StakeAmount))
			statedb.SetState(sfc.ContractAddress, meritPos.DelegatedMe(), utils.BigTo256(it.Staker.DelegatedMe))
			statedb.SetState(sfc.ContractAddress, meritPos.BaseRewardWeight(), utils.BigTo256(baseRewardWeight))
			statedb.SetState(sfc.ContractAddress, meritPos.TxRewardWeight(), utils.BigTo256(txRewardWeight))

			totalBaseRewardWeight.Add(totalBaseRewardWeight, baseRewardWeight)
			totalTxRewardWeight.Add(totalTxRewardWeight, txRewardWeight)
		}
		baseRewardPerSec := s.getRewardPerSec()

		// set total supply
		baseRewards := new(big.Int).Mul(big.NewInt(es.Duration().Unix()), baseRewardPerSec)
		rewards := new(big.Int).Add(baseRewards, epochFee)
		totalSupply := new(big.Int).Add(s.store.app.GetTotalSupply(), rewards)
		statedb.SetState(sfc.ContractAddress, sfcpos.CurrentSealedEpoch(), utils.U64to256(uint64(epoch)))
		s.store.app.SetTotalSupply(totalSupply)

		statedb.SetState(sfc.ContractAddress, epochPos.TotalBaseRewardWeight(), utils.BigTo256(totalBaseRewardWeight))
		statedb.SetState(sfc.ContractAddress, epochPos.TotalTxRewardWeight(), utils.BigTo256(totalTxRewardWeight))
		statedb.SetState(sfc.ContractAddress, epochPos.EpochFee(), utils.BigTo256(epochFee))
		statedb.SetState(sfc.ContractAddress, epochPos.EndTime(), utils.U64to256(uint64(block.Time.Unix())))
		statedb.SetState(sfc.ContractAddress, epochPos.Duration(), utils.U64to256(uint64(es.Duration().Unix())))
		statedb.SetState(sfc.ContractAddress, epochPos.BaseRewardPerSecond(), utils.BigTo256(baseRewardPerSec))
		statedb.SetState(sfc.ContractAddress, epochPos.StakeTotalAmount(), utils.BigTo256(totalStake))
		statedb.SetState(sfc.ContractAddress, epochPos.DelegationsTotalAmount(), utils.BigTo256(totalDelegated))
		statedb.SetState(sfc.ContractAddress, epochPos.TotalSupply(), utils.BigTo256(totalSupply))
		statedb.SetState(sfc.ContractAddress, sfcpos.CurrentSealedEpoch(), utils.U64to256(uint64(epoch)))

		// Add balance for SFC to pay rewards
		statedb.AddBalance(sfc.ContractAddress, rewards)

		// Select new validators
		newStakers := s.GetActiveSfcStakers()
		es.ValidatorProfiles = newStakers

		builder := pos.NewBigBuilder()
		for _, it := range newStakers {
			builder.Set(it.ValidatorID, it.Staker.CalcTotalStake())
		}
		newValidators := builder.Build()
		oldValidators := es.Validators

		// Build new []ValidatorEpochState and []ValidatorBlockState
		newValidatorEpochStates := make([]ValidatorEpochState, newValidators.Len())
		newValidatorBlockStates := make([]ValidatorBlockState, newValidators.Len())
		for newValIdx := 0; newValIdx < newValidators.Len(); newValIdx++ {
			// default values
			newValidatorBlockStates[newValIdx] = ValidatorBlockState{
				Originated: new(big.Int),
			}
			// inherit validator's state from previous epoch, if any
			valID := newValidators.SortedIDs()[newValIdx]
			if !oldValidators.Exists(valID) {
				continue
			}
			oldValIdx := oldValidators.GetIdx(valID)
			newValidatorBlockStates[newValIdx] = bs.ValidatorStates[oldValIdx]
			newValidatorBlockStates[newValIdx].DirtyGasRefund = 0
			newValidatorEpochStates[newValIdx].GasRefund = bs.ValidatorStates[oldValIdx].DirtyGasRefund
			newValidatorEpochStates[newValIdx].PrevEpochEvent = bs.ValidatorStates[oldValIdx].PrevEvent
		}
		es.ValidatorStates = newValidatorEpochStates
		bs.ValidatorStates = newValidatorBlockStates
		es.Validators = newValidators
	}

	if sealEpoch {
		// dirty EpochStats becomes active
		es.PrevEpochStart = es.EpochStart
		es.EpochStart = block.Time
		bs.EpochFee = new(big.Int)
	} else {
		bs.EpochFee = epochFee
	}
}
