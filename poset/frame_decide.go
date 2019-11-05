package poset

import (
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

func (p *Poset) confirmEvents(frame idx.Frame, atropos hash.Event, onEventConfirmed func(*inter.EventHeaderData)) {
	err := p.dfsSubgraph(atropos, func(header *inter.EventHeaderData) bool {
		decidedFrame := p.store.GetEventConfirmedOn(header.Hash())
		if decidedFrame != 0 {
			return false
		}
		// mark all the walked events as confirmed
		p.store.SetEventConfirmedOn(header.Hash(), frame)
		if onEventConfirmed != nil {
			onEventConfirmed(header)
		}
		return true
	})
	if err != nil {
		p.Log.Crit("Poset: Failed to walk subgraph", "err", err)
	}
}

func (p *Poset) confirmBlock(frame idx.Frame, atropos hash.Event, onEventConfirmed func(*inter.EventHeaderData)) (*inter.Block, headersByCreator) {
	lastHeaders := make(headersByCreator, p.Validators.Len())
	blockEvents := make([]*inter.EventHeaderData, 0, int(p.dag.MaxValidatorEventsInBlock)*p.Validators.Len())

	atroposHighestBefore := p.vecClock.GetHighestBeforeAllBranches(atropos)
	validatorIdxs := p.Validators.Idxs()
	var highestLamport idx.Lamport
	var lowestLamport idx.Lamport
	p.confirmEvents(frame, atropos, func(confirmedEvent *inter.EventHeaderData) {
		if onEventConfirmed != nil {
			onEventConfirmed(confirmedEvent)
		}

		// track highest and lowest Lamports
		if highestLamport == 0 || highestLamport < confirmedEvent.Lamport {
			highestLamport = confirmedEvent.Lamport
		}
		if lowestLamport == 0 || lowestLamport > confirmedEvent.Lamport {
			lowestLamport = confirmedEvent.Lamport
		}

		// but not all the events are included into a block
		creatorHighest := atroposHighestBefore.Get(validatorIdxs[confirmedEvent.Creator])
		fromCheater := creatorHighest.IsForkDetected()
		freshEvent := (creatorHighest.Seq - confirmedEvent.Seq) < p.dag.MaxValidatorEventsInBlock // will overflow on forks, it's fine
		if !fromCheater && freshEvent {
			blockEvents = append(blockEvents, confirmedEvent)

			if creatorHighest.Seq == confirmedEvent.Seq {
				lastHeaders[confirmedEvent.Creator] = confirmedEvent
			}
		}
		// sanity check
		if !fromCheater && confirmedEvent.Seq > creatorHighest.Seq {
			p.Log.Crit("DAG is inconsistent with vector clock", "event", confirmedEvent.String(), "seq", confirmedEvent.Seq, "highest", creatorHighest.Seq)
		}
	})

	p.Log.Debug("Confirmed events by", "atropos", atropos.String(), "num", len(blockEvents))

	// ordering
	orderedBlockEvents := p.fareOrdering(blockEvents)
	frameInfo := p.fareTimestamps(frame, atropos, highestLamport, lowestLamport)
	p.store.SetFrameInfo(p.EpochN, frame, &frameInfo)

	// block building
	block := inter.NewBlock(p.Checkpoint.LastBlockN+1, frameInfo.LastConsensusTime, orderedBlockEvents, p.Checkpoint.LastAtropos)
	return block, lastHeaders
}

// onFrameDecided moves LastDecidedFrameN to frame.
// It includes: moving current decided frame, txs ordering and execution, epoch sealing.
func (p *Poset) onFrameDecided(frame idx.Frame, atropos hash.Event) bool {
	p.Log.Debug("consensus: event is atropos", "event", atropos.String())

	p.election.Reset(p.Validators, frame+1)
	p.LastDecidedFrame = frame

	block, lastHeaders := p.confirmBlock(frame, atropos, nil)

	// new checkpoint
	p.Checkpoint.LastBlockN++
	if p.applyBlock != nil {
		p.Checkpoint.StateHash, p.NextValidators = p.applyBlock(block, p.Checkpoint.StateHash, p.NextValidators)
	}
	p.Checkpoint.LastAtropos = atropos
	p.Checkpoint.NextValidators = p.Checkpoint.NextValidators.Top()
	p.saveCheckpoint()

	return p.tryToSealEpoch(atropos, lastHeaders)
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
	p.PrevEpoch.StateHash = p.Checkpoint.StateHash
	p.PrevEpoch.LastHeaders = lastHeaders

	// new validators list, move to new epoch
	p.setEpochValidators(p.NextValidators.Top(), p.EpochN+1)
	p.NextValidators = p.Validators.Copy()
	p.LastDecidedFrame = 0

	// commit
	p.store.SetEpoch(&p.EpochState)
	p.saveCheckpoint()

	// reset internal epoch DB
	p.store.RecreateEpochDb(p.EpochN)

	// reset election & vectorindex to new epoch db
	p.vecClock.Reset(p.Validators, p.store.epochTable.VectorIndex, func(id hash.Event) *inter.EventHeaderData {
		return p.input.GetEventHeader(p.EpochN, id)
	})
	p.election.Reset(p.Validators, firstFrame)

}
