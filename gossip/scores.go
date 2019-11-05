package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// calcScores calculates the validators scores
func (s *Service) calcScores(block *inter.Block, sealEpoch bool) {
	validators := s.engine.GetValidators()

	// Calc validators score
	s.store.SetBlockGasUsed(block.Index, block.GasUsed)
	for v := range validators.Iterate() {
		// Check if validator has confirmed events by this Atropos
		missedBlock := !s.blockParticipated[v]

		// If have confirmed events by this Atropos - just add missed blocks for validator
		if missedBlock {
			s.store.IncBlocksMissed(v)
			continue
		}

		missed := s.store.GetBlocksMissed(v)
		s.store.AddDirtyValidatorsScore(v, block.GasUsed)

		missedCapped := missed
		if missedCapped > uint32(s.config.Net.Economy.FrameLatency) {
			missedCapped = uint32(s.config.Net.Economy.FrameLatency)
		}

		// Add score for previous blocks, but no more than FrameLatency prev blocks
		for i := uint32(1); i <= missedCapped; i++ {
			usedGas := s.store.GetBlockGasUsed(block.Index - idx.Block(i))
			s.store.AddDirtyValidatorsScore(v, usedGas)
		}
		s.store.ResetBlocksMissed(v)
	}

	if sealEpoch {
		lastCheckpoint := s.store.GetScoreCheckpoint()
		if block.Time.Time().Sub(lastCheckpoint.Time()) > s.config.Net.Economy.IntervalBetweenScoreCheckpoints {
			s.store.MoveDirtyValidatorsToActive()
			s.store.SetScoreCheckpoint(block.Time)
		}
	}
}
