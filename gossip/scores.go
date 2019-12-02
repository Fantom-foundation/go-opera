package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type BlocksMissed struct {
	Num    idx.Block
	Period inter.Timestamp
}

// updateOriginationScores calculates the origination scores
func (s *Service) updateOriginationScores(block *inter.Block, receipts types.Receipts, blockEvents inter.Events, txPositions map[common.Hash]TxPosition, sealEpoch bool) {
	// Calc origination scores
	for _, receipt := range receipts {
		txEventPos := txPositions[receipt.TxHash]
		// sanity check
		if txEventPos.Block != block.Index || txEventPos.EventOffset >= uint32(len(block.Events)) {
			s.Log.Crit("Incorrect tx block position", "tx", receipt.TxHash)
		}

		txEvent := blockEvents[txEventPos.BlockOffset]
		// sanity check
		if txEventPos.EventOffset >= uint32(txEvent.Transactions.Len()) || txEvent.Transactions[txEventPos.EventOffset].Hash() == receipt.TxHash {
			s.Log.Crit("Incorrect tx event position", "tx", receipt.TxHash)
		}

		stakerID := idx.StakerID(1) // TODO txEvent.Creator -> StakerID
		s.store.AddDirtyOriginationScore(stakerID, receipt.GasUsed)
	}

	if sealEpoch {
		lastCheckpoint := s.store.GetOriginationScoreCheckpoint()
		if block.Time.Time().Sub(lastCheckpoint.Time()) > s.config.Net.Economy.ScoreCheckpointsInterval {
			s.store.MoveDirtyOriginationScoresToActive()
			s.store.SetOriginationScoreCheckpoint(block.Time)
		}
	}
}

// updateValidationScores calculates the validation scores
func (s *Service) updateValidationScores(block *inter.Block, sealEpoch bool) {
	blockTimeDiff := block.Time - s.store.GetBlock(block.Index-1).Time

	// Calc validation scores
	for _, it := range s.store.GetSfcStakers() {
		// Check if validator has confirmed events by this Atropos
		missedBlock := !s.blockParticipated[it.Staker.Address]

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
		s.store.AddDirtyValidationScore(it.StakerID, block.GasUsed)
		for i := idx.Block(1); i <= missedNum; i++ {
			usedGas := s.store.GetBlock(block.Index - idx.Block(i)).GasUsed
			s.store.AddDirtyValidationScore(it.StakerID, usedGas)
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
