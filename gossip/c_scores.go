package gossip

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/evmcore"
)

const (
	minGasPowerRefund = 800
)

// updateOriginationScores calculates the origination scores
func (s *Service) updateOriginationScores(bs *BlockState, es *EpochState, evmBlock *evmcore.EvmBlock, receipts types.Receipts, txPositions map[common.Hash]TxPosition) {
	// Calc origination scores
	for i, tx := range evmBlock.Transactions {
		txEventPos := txPositions[receipts[i].TxHash]

		txEvent := s.store.GetEvent(txEventPos.Event)
		// sanity check
		if txEvent == nil {
			s.Log.Crit("Incorrect tx event position", "tx", receipts[i].TxHash, "event", txEventPos.Event, "reason", "event has no transactions")
		}

		txFee := new(big.Int).Mul(new(big.Int).SetUint64(receipts[i].GasUsed), tx.GasPrice())

		creatorIdx := es.Validators.GetIdx(txEvent.Creator())
		originated := bs.ValidatorStates[creatorIdx].Originated
		originated.Add(originated, txFee)

		{ // logic for gas power refunds
			if tx.Gas() < receipts[i].GasUsed {
				s.Log.Crit("Transaction gas used is higher than tx gas limit", "tx", receipts[i].TxHash, "event", txEventPos.Event)
			}
			notUsedGas := tx.Gas() - receipts[i].GasUsed
			if notUsedGas >= minGasPowerRefund { // do not refund if refunding is more costly than refunded value
				bs.ValidatorStates[creatorIdx].DirtyGasRefund += notUsedGas
			}
		}
	}
}
