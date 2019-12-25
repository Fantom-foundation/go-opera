package poset

import (
	"math"

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

func (p *Poset) confirmBlock(frame idx.Frame, atropos hash.Event) (block *inter.Block, cheaters []idx.StakerID) {
	blockEvents := make([]*inter.EventHeaderData, 0, 50*p.Validators.Len())

	atroposHighestBefore := p.vecClock.GetHighestBeforeAllBranches(atropos)
	var highestLamport idx.Lamport
	var lowestLamport idx.Lamport
	var confirmedNum int

	// cheaters are ordered deterministically
	cheaters = make([]idx.StakerID, 0, p.Validators.Len())
	for creatorIdx, creator := range p.Validators.SortedIDs() {
		if atroposHighestBefore.Get(idx.Validator(creatorIdx)).IsForkDetected() {
			cheaters = append(cheaters, creator)
		}
	}

	p.confirmEvents(frame, atropos, func(confirmedEvent *inter.EventHeaderData) {
		confirmedNum++

		// track highest and lowest Lamports
		if highestLamport == 0 || highestLamport < confirmedEvent.Lamport {
			highestLamport = confirmedEvent.Lamport
		}
		if lowestLamport == 0 || lowestLamport > confirmedEvent.Lamport {
			lowestLamport = confirmedEvent.Lamport
		}

		// but not all the events are included into a block
		creatorHighest := atroposHighestBefore.Get(p.Validators.GetIdx(confirmedEvent.Creator))
		fromCheater := creatorHighest.IsForkDetected()
		// seqDepth is the depth in of this event in "chain" of self-parents of this creator
		seqDepth := creatorHighest.Seq - confirmedEvent.Seq
		if creatorHighest.Seq < confirmedEvent.Seq {
			seqDepth = math.MaxInt32
		}
		allowed := p.callback.IsEventAllowedIntoBlock == nil || p.callback.IsEventAllowedIntoBlock(confirmedEvent, seqDepth)
		// block consists of allowed events from non-cheaters
		if !fromCheater && allowed {
			blockEvents = append(blockEvents, confirmedEvent)
		}
		// sanity check
		if !fromCheater && confirmedEvent.Seq > creatorHighest.Seq {
			p.Log.Crit("DAG is inconsistent with vector clock", "event", confirmedEvent.String(), "seq", confirmedEvent.Seq, "highest", creatorHighest.Seq)
		}

		if p.callback.OnEventConfirmed != nil {
			p.callback.OnEventConfirmed(confirmedEvent, seqDepth)
		}
	})

	p.Log.Debug("Confirmed events by", "atropos", atropos.String(), "events", confirmedNum, "blocksEvents", len(blockEvents))

	// ordering
	orderedBlockEvents := p.fareOrdering(blockEvents)
	frameInfo := p.fareTimestamps(frame, atropos, highestLamport, lowestLamport)
	p.store.SetFrameInfo(p.EpochN, frame, &frameInfo)

	// block building
	block = inter.NewBlock(p.Checkpoint.LastBlockN+1, frameInfo.LastConsensusTime, atropos, p.Checkpoint.LastAtropos, orderedBlockEvents)
	return block, cheaters
}

// onFrameDecided moves LastDecidedFrameN to frame.
// It includes: moving current decided frame, txs ordering and execution, epoch sealing.
func (p *Poset) onFrameDecided(frame idx.Frame, atropos hash.Event) bool {
	p.Log.Debug("consensus: event is atropos", "event", atropos.String())

	p.election.Reset(p.Validators, frame+1)
	p.Checkpoint.LastDecidedFrame = frame

	block, cheaters := p.confirmBlock(frame, atropos)

	// new checkpoint
	var sealEpoch bool
	var appHash common.Hash
	p.Checkpoint.LastBlockN++
	if p.callback.ApplyBlock != nil {
		appHash, sealEpoch = p.callback.ApplyBlock(block, frame, cheaters)
		p.Checkpoint.AppHash = hash.Of(p.Checkpoint.AppHash.Bytes(), appHash.Bytes())
	}
	p.Checkpoint.LastAtropos = atropos
	p.saveCheckpoint()

	if sealEpoch {
		p.sealEpoch()
	}
	return sealEpoch
}

func (p *Poset) sealEpoch() {
	// new PrevEpoch state
	p.PrevEpoch.Time = p.frameConsensusTime(p.LastDecidedFrame)
	p.PrevEpoch.Epoch = p.EpochN
	p.PrevEpoch.LastAtropos = p.Checkpoint.LastAtropos
	p.PrevEpoch.AppHash = p.Checkpoint.AppHash

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
