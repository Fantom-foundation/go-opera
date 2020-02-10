package gossip

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

// PoiPeriod calculate POI period from int64 unix time
func PoiPeriod(t inter.Timestamp, config *lachesis.EconomyConfig) uint64 {
	return uint64(t) / uint64(config.PoiPeriodDuration)
}

// UpdateAddressPOI calculate and save POI for user
func (s *Service) UpdateAddressPOI(address common.Address, senderTotalFee *big.Int, poiPeriod uint64) {
	/*if senderTotalFee.Sign() == 0 {
		s.store.SetAddressPOI(address, common.Big0)
		return // avoid division by 0
	}
	poi := new(big.Int).Mul(senderTotalFee, lachesis.PercentUnit)
	poi.Div(poi, s.store.GetPoiFee(poiPeriod)) // rebase user's PoI as <= 1.0 ratio
	s.store.SetAddressPOI(address, poi)*/
}

// updateUsersPOI calculates the Proof Of Importance weights for users
func (s *Service) updateUsersPOI(block *inter.Block, evmBlock *evmcore.EvmBlock, receipts types.Receipts, totalFee *big.Int, sealEpoch bool) {
	// User POI calculations
	poiPeriod := PoiPeriod(block.Time, &s.config.Net.Economy)
	s.app.AddPoiFee(poiPeriod, totalFee)

	for i, tx := range evmBlock.Transactions {
		txFee := new(big.Int).Mul(new(big.Int).SetUint64(receipts[i].GasUsed), tx.GasPrice())

		signer := types.NewEIP155Signer(s.config.Net.EvmChainConfig().ChainID)
		sender, err := signer.Sender(tx)
		if err != nil {
			s.Log.Crit("Failed to get sender from transaction", "err", err)
		}

		senderLastTxTime := s.app.GetAddressLastTxTime(sender)
		prevUserPoiPeriod := PoiPeriod(senderLastTxTime, &s.config.Net.Economy)
		senderTotalFee := s.app.GetAddressFee(sender, prevUserPoiPeriod)

		delegator := s.app.GetSfcDelegator(sender)
		if delegator != nil {
			staker := s.app.GetSfcStaker(delegator.ToStakerID)
			if staker != nil {
				prevWeightedTxFee := s.app.GetWeightedDelegatorsFee(delegator.ToStakerID)

				weightedTxFee := new(big.Int).Mul(txFee, delegator.Amount)
				weightedTxFee.Div(weightedTxFee, staker.CalcTotalStake())

				weightedTxFee.Add(weightedTxFee, prevWeightedTxFee)
				s.app.SetWeightedDelegatorsFee(delegator.ToStakerID, weightedTxFee)
			}
		}

		if prevUserPoiPeriod != poiPeriod {
			s.UpdateAddressPOI(sender, senderTotalFee, prevUserPoiPeriod)
			senderTotalFee = big.NewInt(0)
		}

		s.app.SetAddressLastTxTime(sender, block.Time)
		senderTotalFee.Add(senderTotalFee, txFee)
		s.app.SetAddressFee(sender, poiPeriod, senderTotalFee)
	}

}

// UpdateStakerPOI calculate and save POI for staker
func (s *Service) UpdateStakerPOI(stakerID idx.StakerID, stakerAddress common.Address, poiPeriod uint64) {
	staker := s.app.GetSfcStaker(stakerID)

	vFee := s.app.GetAddressFee(stakerAddress, poiPeriod)
	weightedDFee := s.app.GetWeightedDelegatorsFee(stakerID)
	if vFee.Sign() == 0 && weightedDFee.Sign() == 0 {
		s.app.SetStakerPOI(stakerID, common.Big0)
		return // optimization
	}

	weightedVFee := new(big.Int).Mul(vFee, staker.StakeAmount)
	weightedVFee.Div(weightedVFee, staker.CalcTotalStake())

	weightedFee := new(big.Int).Add(weightedDFee, weightedVFee)

	if weightedFee.Sign() == 0 {
		s.app.SetStakerPOI(stakerID, common.Big0)
		return // avoid division by 0
	}
	poi := weightedFee // no need to rebase validator's PoI as <= 1.0 ratio
	/*poi := new(big.Int).Mul(weightedFee, lachesis.PercentUnit)
	poi.Div(poi, s.store.GetPoiFee(poiPeriod))*/
	s.app.SetStakerPOI(stakerID, poi)
}

// updateStakersPOI calculates the Proof Of Importance weights for stakers
func (s *Service) updateStakersPOI(block *inter.Block, sealEpoch bool) {
	// Stakers POI calculations
	poiPeriod := PoiPeriod(block.Time, &s.config.Net.Economy)
	prevBlockPoiPeriod := PoiPeriod(s.store.GetBlock(block.Index-1).Time, &s.config.Net.Economy)

	if poiPeriod != prevBlockPoiPeriod {
		for _, it := range s.GetActiveSfcStakers() {
			s.UpdateStakerPOI(it.StakerID, it.Staker.Address, prevBlockPoiPeriod)
		}
		// clear StakersDelegatorsFee counters
		s.app.DelAllWeightedDelegatorsFee()
	}
}
