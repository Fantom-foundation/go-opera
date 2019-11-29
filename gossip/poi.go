package gossip

import (
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
func (s *Store) UpdateAddressPOI(address common.Address, senderTotalGasUsed uint64, poiPeriod uint64) {
	if senderTotalGasUsed == 0 {
		s.SetAddressPOI(address, 0)
	} else {
		poi := uint64((senderTotalGasUsed * 1000000) / s.GetPOIGasUsed(poiPeriod))
		s.SetAddressPOI(address, poi)
	}
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
			prevGas := s.store.GetStakerDelegatorsGasUsed(delegator.ToStakerID)
			s.store.SetStakerDelegatorsGasUsed(delegator.ToStakerID, prevGas+txGasUsed)
		}

		if prevUserPoiPeriod != poiPeriod {
			s.store.UpdateAddressPOI(sender, senderTotalGasUsed, prevUserPoiPeriod)
			senderTotalGasUsed = 0
		}

		s.store.SetAddressLastTxTime(sender, block.Time)
		senderTotalGasUsed += txGasUsed
		s.store.SetAddressGasUsed(sender, poiPeriod, senderTotalGasUsed)
	}

}

// UpdateStakerPOI calculate and save POI for staker
func (s *Store) UpdateStakerPOI(stakerID idx.StakerID, stakerAddress common.Address, poiPeriod uint64) {
	dGasUsed := s.GetStakerDelegatorsGasUsed(stakerID)
	vGasUsed := s.GetAddressGasUsed(stakerAddress, poiPeriod)

	vGasUsed += dGasUsed

	if vGasUsed == 0 {
		s.SetStakerPOI(stakerID, 0)
	} else {
		poi := (vGasUsed * 1000000) / s.GetPOIGasUsed(poiPeriod)
		s.SetStakerPOI(stakerID, poi)
	}
}

// updateStakersPOI calculates the Proof Of Importance weights for stakers
func (s *Service) updateStakersPOI(block *inter.Block, sealEpoch bool) {
	// Stakers POI calculations
	poiPeriod := PoiPeriod(block.Time, &s.config.Net.Economy)
	prevBlockPoiPeriod := PoiPeriod(s.store.GetBlock(block.Index-1).Time, &s.config.Net.Economy)

	if poiPeriod != prevBlockPoiPeriod {
		for _, it := range s.store.GetSfcStakers() {
			s.store.UpdateStakerPOI(it.StakerID, it.Staker.Address, prevBlockPoiPeriod)
		}
		// clear StakersDelegatorsGasUsed counters
		s.store.DelStakersDelegatorsGasUsed()
	}
}
