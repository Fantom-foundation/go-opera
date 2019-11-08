package poset

import (
	"github.com/ethereum/go-ethereum/common"

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

func (p *Poset) confirmBlock(frame idx.Frame, atropos hash.Event) (block *inter.Block, cheaters []common.Address, lastHeaders headersByCreator) {
	lastHeaders = make(headersByCreator, p.Validators.Len())
	blockEvents := make([]*inter.EventHeaderData, 0, int(p.dag.MaxValidatorEventsInBlock)*p.Validators.Len())

	atroposHighestBefore := p.vecClock.GetHighestBeforeAllBranches(atropos)
	validatorIdxs := p.Validators.Idxs()
	var highestLamport idx.Lamport
	var lowestLamport idx.Lamport
	var confirmedNum int

	cheaters = make([]common.Address, 0, len(validatorIdxs))
	for creator, creatorIdx := range validatorIdxs {
		if atroposHighestBefore.Get(creatorIdx).IsForkDetected() {
			cheaters = append(cheaters, creator)
		}
	}

	p.confirmEvents(frame, atropos, func(confirmedEvent *inter.EventHeaderData) {
		if p.callback.OnEventConfirmed != nil {
			p.callback.OnEventConfirmed(confirmedEvent)
		}
		confirmedNum++

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
		allowed := p.callback.IsEventAllowedIntoBlock == nil || p.callback.IsEventAllowedIntoBlock(confirmedEvent, creatorHighest.Seq)
		// block consists of allowed events from non-cheaters
		if !fromCheater && allowed {
			blockEvents = append(blockEvents, confirmedEvent)
		}
		// track last confirmed events from each validator
		if !fromCheater && creatorHighest.Seq == confirmedEvent.Seq {
			lastHeaders[confirmedEvent.Creator] = confirmedEvent
		}
		// sanity check
		if !fromCheater && confirmedEvent.Seq > creatorHighest.Seq {
			p.Log.Crit("DAG is inconsistent with vector clock", "event", confirmedEvent.String(), "seq", confirmedEvent.Seq, "highest", creatorHighest.Seq)
		}
	})

	p.Log.Debug("Confirmed events by", "atropos", atropos.String(), "events", confirmedNum, "blocksEvents", len(blockEvents))

	// ordering
	orderedBlockEvents := p.fareOrdering(blockEvents)
	frameInfo := p.fareTimestamps(frame, atropos, highestLamport, lowestLamport)
	p.store.SetFrameInfo(p.EpochN, frame, &frameInfo)

	// block building
	block = inter.NewBlock(p.Checkpoint.LastBlockN+1, frameInfo.LastConsensusTime, atropos, p.Checkpoint.LastAtropos, orderedBlockEvents)
	return block, cheaters, lastHeaders
}

// onFrameDecided moves LastDecidedFrameN to frame.
// It includes: moving current decided frame, txs ordering and execution, epoch sealing.
func (p *Poset) onFrameDecided(frame idx.Frame, atropos hash.Event) bool {
	p.Log.Debug("consensus: event is atropos", "event", atropos.String())

	p.election.Reset(p.Validators, frame+1)
	p.Checkpoint.LastDecidedFrame = frame

	block, cheaters, lastHeaders := p.confirmBlock(frame, atropos)

	// new checkpoint
	var sealEpoch bool
	p.Checkpoint.LastBlockN++
	if p.callback.ApplyBlock != nil {
		p.Checkpoint.StateHash, sealEpoch = p.callback.ApplyBlock(inter.ApplyBlockArgs{
			Block:        block,
			DecidedFrame: frame,
			StateHash:    p.Checkpoint.StateHash,
			Validators:   p.Validators,
			Cheaters:     cheaters,
		})
	}
	p.Checkpoint.LastAtropos = atropos
	p.saveCheckpoint()

	if sealEpoch {
		p.sealEpoch(lastHeaders)
	}
	return sealEpoch
}

func (p *Poset) sealEpoch(lastHeaders headersByCreator) {
	// new PrevEpoch state
	p.PrevEpoch.Time = p.frameConsensusTime(p.LastDecidedFrame)
	p.PrevEpoch.Epoch = p.EpochN
	p.PrevEpoch.LastAtropos = p.Checkpoint.LastAtropos
	p.PrevEpoch.StateHash = p.Checkpoint.StateHash
	p.PrevEpoch.LastHeaders = lastHeaders

	// new validators list, move to new epoch
	nextValidators := p.Validators
	if p.callback.SelectValidatorsGroup != nil {
		nextValidators = p.callback.SelectValidatorsGroup(p.EpochN, p.EpochN+1)
	}
	p.setEpochValidators(nextValidators, p.EpochN+1)
	p.Checkpoint.LastDecidedFrame = 0

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
