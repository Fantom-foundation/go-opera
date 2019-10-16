package poset

import (
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

func (p *Poset) confirmBlockEvents(frame idx.Frame, atropos hash.Event) ([]*inter.EventHeaderData, headersByCreator) {
	lastHeaders := make(headersByCreator, len(p.Validators))
	blockEvents := make([]*inter.EventHeaderData, 0, int(p.dag.MaxValidatorEventsInBlock)*len(p.Validators))

	atroposHighestBefore := p.vecClock.GetHighestBeforeSeq(atropos)
	validatorIdxs := p.Validators.Idxs()
	err := p.dfsSubgraph(atropos, func(header *inter.EventHeaderData) bool {
		decidedFrame := p.store.GetEventConfirmedOn(header.Hash())
		if decidedFrame != 0 {
			return false
		}
		// mark all the walked events
		p.store.SetEventConfirmedOn(header.Hash(), frame)
		// but not all the events are included into a block
		creatorHighest := atroposHighestBefore.Get(validatorIdxs[header.Creator])
		fromCheater := creatorHighest.IsForkDetected()
		freshEvent := (creatorHighest.Seq - header.Seq) < p.dag.MaxValidatorEventsInBlock // will overflow on forks, it's fine
		if !fromCheater && freshEvent {
			blockEvents = append(blockEvents, header)

			if creatorHighest.Seq == header.Seq {
				lastHeaders[header.Creator] = header
			}
		}
		// sanity check
		if !fromCheater && header.Seq > creatorHighest.Seq {
			p.Log.Crit("DAG is inconsistent with vector clock", "event", header.String(), "seq", header.Seq, "highest", creatorHighest.Seq)
		}
		return true
	})
	if err != nil {
		p.Log.Crit("Poset: Failed to walk subgraph", "err", err)
	}

	p.Log.Debug("Confirmed events by", "atropos", atropos.String(), "num", len(blockEvents))
	return blockEvents, lastHeaders
}

// onFrameDecided moves LastDecidedFrameN to frame.
// It includes: moving current decided frame, txs ordering and execution, epoch sealing.
func (p *Poset) onFrameDecided(frame idx.Frame, atropos hash.Event) headersByCreator {
	p.Log.Debug("consensus: event is atropos", "event", atropos.String())

	p.election.Reset(p.Validators, frame+1)
	p.LastDecidedFrame = frame

	blockEvents, lastHeaders := p.confirmBlockEvents(frame, atropos)

	// ordering
	if len(blockEvents) == 0 {
		p.Log.Crit("Frame is decided with no events. It isn't possible.")
	}
	ordered, frameInfo := p.fareOrdering(frame, atropos, blockEvents)

	// block generation
	p.checkpoint.LastBlockN++
	if p.applyBlock != nil {
		block := inter.NewBlock(p.checkpoint.LastBlockN, frameInfo.LastConsensusTime, ordered, p.checkpoint.LastAtropos)
		p.checkpoint.StateHash, p.NextValidators = p.applyBlock(block, p.checkpoint.StateHash, p.NextValidators)
	}
	p.checkpoint.LastAtropos = atropos
	p.NextValidators = p.NextValidators.Top()

	p.saveCheckpoint()

	return lastHeaders
}

func (p *Poset) isEpochSealed() bool {
	return p.LastDecidedFrame >= p.dag.EpochLen
}

func (p *Poset) tryToSealEpoch(atropos hash.Event, lastHeaders headersByCreator) bool {
	if !p.isEpochSealed() {
		return false
	}

	p.onNewEpoch(atropos, lastHeaders)

	return true
}

func (p *Poset) onNewEpoch(atropos hash.Event, lastHeaders headersByCreator) {
	// new PrevEpoch state
	p.PrevEpoch.Time = p.frameConsensusTime(p.LastDecidedFrame)
	p.PrevEpoch.Epoch = p.EpochN
	p.PrevEpoch.LastAtropos = atropos
	p.PrevEpoch.StateHash = p.checkpoint.StateHash
	p.PrevEpoch.LastHeaders = lastHeaders

	// new validators list
	p.Validators = p.NextValidators.Top()
	p.NextValidators = p.Validators.Copy()

	// move to new epoch
	p.EpochN++
	p.LastDecidedFrame = 0

	// commit
	p.store.SetEpoch(&p.epochState)
	p.saveCheckpoint()

	// reset internal epoch DB
	p.store.RecreateEpochDb(p.EpochN)

	// reset election & vectorindex to new epoch db
	p.vecClock.Reset(p.Validators, p.store.epochTable.VectorIndex, func(id hash.Event) *inter.EventHeaderData {
		return p.input.GetEventHeader(p.EpochN, id)
	})
	p.election.Reset(p.Validators, firstFrame)

}
