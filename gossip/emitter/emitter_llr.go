package emitter

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/go-opera/eventcheck/basiccheck"
	"github.com/Fantom-foundation/go-opera/inter"
)

func (em *Emitter) addLlrBlockVotes(e *inter.MutableEventPayload) {
	if e.Version() == 0 {
		e.SetBlockVotes(inter.LlrBlockVotes{
			Votes: []hash.Hash{},
		})
		return
	}
	start := em.world.GetLowestBlockToDecide()
	prevInDB := em.world.GetLastBV(e.Creator())
	if prevInDB != nil && start < *prevInDB+1 {
		start = *prevInDB + 1
	}
	prevInFile := em.readLastBlockVotes()
	if prevInFile != nil && start < *prevInFile+1 {
		start = *prevInFile + 1
	}
	records := make([]hash.Hash, 0, 16)
	var epoch idx.Epoch
	for b := start; len(records) < basiccheck.MaxBlockVotesPerEvent; b++ {
		record := em.world.GetBlockRecordHash(b)
		if record == nil {
			break
		}
		blockEpoch := em.world.GetBlockEpoch(b)
		if epoch == 0 {
			epoch = blockEpoch
		}
		if epoch != blockEpoch || blockEpoch == 0 {
			break
		}
		records = append(records, *record)
	}
	if len(records) == 0 {
		start = 0
		epoch = 0
	}
	e.SetBlockVotes(inter.LlrBlockVotes{
		Start: start,
		Epoch: epoch,
		Votes: records,
	})
}

func (em *Emitter) addLlrEpochVote(e *inter.MutableEventPayload) {
	if e.Version() == 0 {
		return
	}
	target := em.world.GetLowestEpochToDecide()
	prevInDB := em.world.GetLastEV(e.Creator())
	if prevInDB != nil && target < *prevInDB+1 {
		target = *prevInDB + 1
	}
	prevInFile := em.readLastEpochVote()
	if prevInFile != nil && target < *prevInFile+1 {
		target = *prevInFile + 1
	}
	vote := em.world.GetEpochRecordHash(target)
	if vote == nil {
		return
	}
	e.SetEpochVote(inter.LlrEpochVote{
		Epoch: target,
		Vote:  *vote,
	})
}
