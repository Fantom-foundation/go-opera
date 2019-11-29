package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// calcScores calculates the validators scores
func (s *Service) updateScores(block *inter.Block, sealEpoch bool) {
	// Calc validators score
	for _, it := range s.store.GetSfcStakers() {
		// Check if validator has confirmed events by this Atropos
		missedBlock := !s.blockParticipated[it.Staker.Address]

		// If have no confirmed events by this Atropos - just add missed blocks for validator
		if missedBlock {
			s.store.IncBlocksMissed(it.StakerID)
			continue
		}

		missed := s.store.GetBlocksMissed(it.StakerID)
		if missed > uint32(s.config.Net.Economy.FrameLatency) {
			missed = uint32(s.config.Net.Economy.FrameLatency)
		}

		// Add score for previous blocks, but no more than FrameLatency prev blocks
		s.store.AddDirtyValidationScore(it.StakerID, block.GasUsed)
		for i := uint32(1); i <= missed; i++ {
			usedGas := s.store.GetBlock(block.Index - idx.Block(i)).GasUsed
			s.store.AddDirtyValidationScore(it.StakerID, usedGas)
		}
		s.store.ResetBlocksMissed(it.StakerID)
	}

	if sealEpoch {
		lastCheckpoint := s.store.GetScoreCheckpoint()
		if block.Time.Time().Sub(lastCheckpoint.Time()) > s.config.Net.Economy.ScoreCheckpointsInterval {
			s.store.MoveDirtyValidatorsToActive()
			s.store.SetScoreCheckpoint(block.Time)
		}
	}
}
