package gossip

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/inter"
)

// calcPOI calculates the Proof Of Importance weights
func (s *Service) calcPOI(block *inter.Block, evmBlock *evmcore.EvmBlock, sealEpoch bool) {
	// POI calculations
	poiPeriod := PoiPeriod(block.Time.Unix())
	s.store.AddPOIGasUsed(poiPeriod, block.GasUsed)

	for _, tx := range evmBlock.Transactions {
		signer := types.NewEIP155Signer(params.AllEthashProtocolChanges.ChainID)
		sender, err := signer.Sender(tx)
		if err != nil {
			s.Log.Crit("Failed to get sender from transaction", "err", err)
		}

		senderTotalGasUsed := s.store.GetAddressGasUsed(sender)
		senderLastTxTime := s.store.GetAddressLastTxTime(sender)
		prevPoiPeriod := PoiPeriod(int64(senderLastTxTime))

		if prevPoiPeriod != poiPeriod {
			delegator := s.store.GetSfcDelegator(sender)
			if delegator != nil {
				s.store.CalcValidatorsPOI(delegator.ToStakerID, sender, prevPoiPeriod)
			}
			s.store.SetAddressGasUsed(sender, 0)
		}

		s.store.SetAddressLastTxTime(sender, uint64(block.Time.Unix()))
		s.store.SetAddressGasUsed(sender, senderTotalGasUsed+tx.Gas())
	}

}
