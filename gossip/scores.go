package gossip

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// BlocksMissed is information about missed blocks from a staker
type BlocksMissed struct {
	Num    idx.Block
	Period inter.Timestamp
}

// updateOriginationScores calculates the origination scores
func (s *Service) updateOriginationScores(block *inter.Block, evmBlock *evmcore.EvmBlock, receipts types.Receipts, txPositions map[common.Hash]TxPosition, sealEpoch bool) {
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

		s.store.AddDirtyOriginationScore(txEvent.Creator, txFee)
	}

	if sealEpoch {
		lastCheckpoint := s.store.GetOriginationScoreCheckpoint()
		if block.Time.Time().Sub(lastCheckpoint.Time()) > s.config.Net.Economy.ScoreCheckpointsInterval {
			s.store.DelAllActiveOriginationScores()
			s.store.MoveDirtyOriginationScoresToActive()
			s.store.SetOriginationScoreCheckpoint(block.Time)
		}
	}
}

// updateValidationScores calculates the validation scores
func (s *Service) updateValidationScores(block *inter.Block, totalFee *big.Int, sealEpoch bool) {
	blockTimeDiff := block.Time - s.store.GetBlock(block.Index-1).Time

	// Write block fee
	s.store.SetBlockFee(block.Index, totalFee)

	// Calc validation scores
	for _, it := range s.GetActiveSfcStakers() {
		// Check if validator has confirmed events by this Atropos
		missedBlock := !s.blockParticipated[it.StakerID]

		// If have no confirmed events by this Atropos - just add missed blocks for validator
		if missedBlock {
			s.store.IncBlocksMissed(it.StakerID, blockTimeDiff)
			continue
		}

		missedNum := s.store.GetBlocksMissed(it.StakerID).Num
		if missedNum > s.config.Net.Economy.BlockMissedLatency {
			missedNum = s.config.Net.Economy.BlockMissedLatency
		}

		// Add score for previous blocks, but no more than FrameLatency prev blocks
		s.store.AddDirtyValidationScore(it.StakerID, totalFee)
		for i := idx.Block(1); i <= missedNum && i < block.Index; i++ {
			fee := s.store.GetBlockFee(block.Index - i)
			s.store.AddDirtyValidationScore(it.StakerID, fee)
		}
		s.store.ResetBlocksMissed(it.StakerID)
	}

	if sealEpoch {
		lastCheckpoint := s.store.GetValidationScoreCheckpoint()
		if block.Time.Time().Sub(lastCheckpoint.Time()) > s.config.Net.Economy.ScoreCheckpointsInterval {
			s.store.MoveDirtyValidationScoresToActive()
			s.store.SetValidationScoreCheckpoint(block.Time)
		}
	}
}
