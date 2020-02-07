package gossip

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

const (
	minGasPowerRefund = 800
)

// updateOriginationScores calculates the origination scores
func (s *Service) updateOriginationScores(block *inter.Block, evmBlock *evmcore.EvmBlock, receipts types.Receipts, txPositions map[common.Hash]TxPosition, sealEpoch bool) {
	epoch := s.engine.GetEpoch()
	// Calc origination scores
	for i, tx := range evmBlock.Transactions {
		txEventPos := txPositions[receipts[i].TxHash]
		// sanity check
		if txEventPos.Block != block.Index {
			s.Log.Crit("Incorrect tx block position", "tx", receipts[i].TxHash,
				"block", txEventPos.Block, "block_got", block.Index)
		}

		txEvent := s.store.GetEventHeader(txEventPos.Event.Epoch(), txEventPos.Event)
		// sanity check
		if txEvent == nil {
			s.Log.Crit("Incorrect tx event position", "tx", receipts[i].TxHash, "event", txEventPos.Event, "reason", "event has no transactions")
		}

		txFee := new(big.Int).Mul(new(big.Int).SetUint64(receipts[i].GasUsed), tx.GasPrice())

		s.app.AddDirtyOriginationScore(txEvent.Creator, txFee)

		{ // logic for gas power refunds
			if tx.Gas() < receipts[i].GasUsed {
				s.Log.Crit("Transaction gas used is higher than tx gas limit", "tx", receipts[i].TxHash, "event", txEventPos.Event)
			}
			notUsedGas := tx.Gas() - receipts[i].GasUsed
			if notUsedGas >= minGasPowerRefund { // do not refund if refunding is more costly than refunded value
				s.app.IncGasPowerRefund(epoch, txEvent.Creator, notUsedGas)
			}
		}
	}

	if sealEpoch {
		s.app.DelAllActiveOriginationScores()
		s.app.MoveDirtyOriginationScoresToActive()
		// prune not needed gas power records
		s.app.DelGasPowerRefunds(epoch - 1)
	}
}

// updateValidationScores calculates the validation scores
func (s *Service) updateValidationScores(block *inter.Block, sealEpoch bool) {
	blockTimeDiff := block.Time - s.store.GetBlock(block.Index-1).Time

	// Calc validation scores
	for _, it := range s.GetActiveSfcStakers() {
		// validators only
		if !s.engine.GetValidators().Exists(it.StakerID) {
			continue
		}

		// Check if validator has confirmed events by this Atropos
		missedBlock := !s.blockParticipated[it.StakerID]

		// If have no confirmed events by this Atropos - just add missed blocks for validator
		if missedBlock {
			s.app.IncBlocksMissed(it.StakerID, blockTimeDiff)
			continue
		}

		missedNum := s.app.GetBlocksMissed(it.StakerID).Num
		if missedNum > s.config.Net.Economy.BlockMissedLatency {
			missedNum = s.config.Net.Economy.BlockMissedLatency
		}

		// Add score for previous blocks, but no more than FrameLatency prev blocks
		s.app.AddDirtyValidationScore(it.StakerID, new(big.Int).SetUint64(uint64(blockTimeDiff)))
		for i := idx.Block(1); i <= missedNum && i < block.Index; i++ {
			blockTime := s.store.GetBlock(block.Index - i).Time
			prevBlockTime := s.store.GetBlock(block.Index - i - 1).Time
			timeDiff := blockTime - prevBlockTime
			s.app.AddDirtyValidationScore(it.StakerID, new(big.Int).SetUint64(uint64(timeDiff)))
		}
		s.app.ResetBlocksMissed(it.StakerID)
	}

	if sealEpoch {
		s.app.DelAllActiveValidationScores()
		s.app.MoveDirtyValidationScoresToActive()
	}
}
