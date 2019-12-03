package gossip

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

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
func (s *Service) UpdateAddressPOI(address common.Address, senderTotalGasUsed uint64, poiPeriod uint64) {
	if senderTotalGasUsed == 0 {
		s.store.SetAddressPOI(address, 0)
		return // avoid division by 0
	}
	poi := (senderTotalGasUsed * 1000000) / s.store.GetPOIGasUsed(poiPeriod)
	s.store.SetAddressPOI(address, poi)
}

// updateUsersPOI calculates the Proof Of Importance weights for users
func (s *Service) updateUsersPOI(block *inter.Block, evmBlock *evmcore.EvmBlock, receipts types.Receipts, sealEpoch bool) {
	// User POI calculations
	poiPeriod := PoiPeriod(block.Time, &s.config.Net.Economy)
	s.store.AddPOIGasUsed(poiPeriod, block.GasUsed)

	for i, tx := range evmBlock.Transactions {
		txGasUsed := receipts[i].GasUsed

		signer := types.NewEIP155Signer(params.AllEthashProtocolChanges.ChainID)
		sender, err := signer.Sender(tx)
		if err != nil {
			s.Log.Crit("Failed to get sender from transaction", "err", err)
		}

		senderLastTxTime := s.store.GetAddressLastTxTime(sender)
		prevUserPoiPeriod := PoiPeriod(senderLastTxTime, &s.config.Net.Economy)
		senderTotalGasUsed := s.store.GetAddressGasUsed(sender, prevUserPoiPeriod)

		delegator := s.store.GetSfcDelegator(sender)
		if delegator != nil {
			staker := s.store.GetSfcStaker(delegator.ToStakerID)
			prevGas := s.store.GetWeightedDelegatorsGasUsed(delegator.ToStakerID)

			weightedTxGasUsed := new(big.Int).SetUint64(txGasUsed)
			weightedTxGasUsed.Mul(weightedTxGasUsed, delegator.Amount)
			weightedTxGasUsed.Div(weightedTxGasUsed, staker.CalcEfficientStake())

			s.store.SetWeightedDelegatorsGasUsed(delegator.ToStakerID, prevGas+weightedTxGasUsed.Uint64())
		}

		if prevUserPoiPeriod != poiPeriod {
			s.UpdateAddressPOI(sender, senderTotalGasUsed, prevUserPoiPeriod)
			senderTotalGasUsed = 0
		}

		s.store.SetAddressLastTxTime(sender, block.Time)
		senderTotalGasUsed += txGasUsed
		s.store.SetAddressGasUsed(sender, poiPeriod, senderTotalGasUsed)
	}

}

// UpdateStakerPOI calculate and save POI for staker
func (s *Service) UpdateStakerPOI(stakerID idx.StakerID, stakerAddress common.Address, poiPeriod uint64) {
	staker := s.store.GetSfcStaker(stakerID)

	vGasUsed := s.store.GetAddressGasUsed(stakerAddress, poiPeriod)
	weightedDGasUsed := s.store.GetWeightedDelegatorsGasUsed(stakerID)
	if vGasUsed == 0 && weightedDGasUsed == 0 {
		s.store.SetStakerPOI(stakerID, 0)
		return // optimization
	}

	weightedVGasUsed := new(big.Int).SetUint64(vGasUsed)
	weightedVGasUsed.Mul(weightedVGasUsed, staker.StakeAmount)
	weightedVGasUsed.Div(weightedVGasUsed, staker.CalcEfficientStake())

	weightedGasUsed := weightedDGasUsed + weightedVGasUsed.Uint64()

	if weightedGasUsed == 0 {
		s.store.SetStakerPOI(stakerID, 0)
		return // avoid division by 0
	}
	poi := (weightedGasUsed * 1000000) / s.store.GetPOIGasUsed(poiPeriod)
	s.store.SetStakerPOI(stakerID, poi)
}

// updateStakersPOI calculates the Proof Of Importance weights for stakers
func (s *Service) updateStakersPOI(block *inter.Block, sealEpoch bool) {
	// Stakers POI calculations
	poiPeriod := PoiPeriod(block.Time, &s.config.Net.Economy)
	prevBlockPoiPeriod := PoiPeriod(s.store.GetBlock(block.Index-1).Time, &s.config.Net.Economy)

	if poiPeriod != prevBlockPoiPeriod {
		for _, it := range s.store.GetSfcStakers() {
			s.UpdateStakerPOI(it.StakerID, it.Staker.Address, prevBlockPoiPeriod)
		}
		// clear StakersDelegatorsGasUsed counters
		s.store.DelAllWeightedDelegatorsGasUsed()
	}
}
