package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"time"
)

const (
	PeriodBetweenScoreCheckpoints = 4 * time.Hour
)

// calcScores calculates the validators scores
func (s *Service) calcScores(block *inter.Block, sealEpoch bool) {
	validators := s.engine.GetValidators()

	// Calc validators score
	s.store.SetBlockGasUsed(block.Index, block.GasUsed)
	for v := range validators.Iterate() {
		// Check validator events in current block
		eventsInBlock := false
		for _, evHash := range block.Events {
			evh := s.store.GetEventHeader(evHash.Epoch(), evHash)
			if evh.Creator == v {
				eventsInBlock = true
				break
			}
		}

		// If have not events in block - add missed blocks for validator
		if !eventsInBlock {
			s.store.IncBlocksMissed(v)
			continue
		}

		missed := s.store.GetBlocksMissed(v)
		s.store.AddDirtyValidatorsScore(v, block.GasUsed)

		missedCapped := missed
		if missedCapped > 2 {
			missedCapped = 2
		}

		for i := uint32(1); i < missedCapped; i++ {
			usedGas := s.store.GetBlockGasUsed(block.Index - idx.Block(i))
			s.store.AddDirtyValidatorsScore(v, usedGas)
		}
		s.store.ResetBlocksMissed(v)
	}

	if sealEpoch {
		prevBlock := s.store.GetBlockByHash(block.PrevHash)
		if len(prevBlock.Events) > 0 && len(block.Events) > 0 {
			if prevBlock.Events[0].Epoch() != block.Events[0].Epoch() {
				// Epoch changed
				lastCheckpoint := s.store.GetScoreCheckpoint()
				if block.Time.Time().Sub(lastCheckpoint.Time()) > PeriodBetweenScoreCheckpoints {
					s.store.MoveDirtyValidatorsToActive()
					s.store.SetScoreCheckpoint(block.Time)
				}
			}
		}
	}
}
