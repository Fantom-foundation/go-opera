package emitter

import (
	"math/rand"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/utils/piecefunc"

	"github.com/Fantom-foundation/go-opera/eventcheck/basiccheck"
	"github.com/Fantom-foundation/go-opera/inter"
)

var emptyLlrBlockVotes = inter.LlrBlockVotes{
	Votes: []hash.Hash{},
}

func (em *Emitter) addLlrBlockVotes(e *inter.MutableEventPayload) {
	if em.skipLlrBlockVote() || e.Version() == 0 {
		e.SetBlockVotes(emptyLlrBlockVotes)
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
	records := make([]hash.Hash, 0, basiccheck.MaxBlockVotesPerEvent)
	epochEnd := false
	var epoch idx.Epoch
	for b := start; len(records) < basiccheck.MaxBlockVotesPerEvent; b++ {
		blockEpoch := em.world.GetBlockEpoch(b)
		if epoch == 0 {
			if !em.isEpochValidator(blockEpoch) {
				continue
			}
			epoch = blockEpoch
			start = b
		}
		if epoch != blockEpoch || blockEpoch == 0 {
			epochEnd = true
			break
		}
		record := em.world.GetBlockRecordHash(b)
		if record == nil {
			break
		}
		records = append(records, *record)
	}

	waitUntilLongerBatch := !epochEnd && len(records) < basiccheck.MaxBlockVotesPerEvent
	if len(records) == 0 || waitUntilLongerBatch {
		e.SetBlockVotes(emptyLlrBlockVotes)
		return
	}
	e.SetBlockVotes(inter.LlrBlockVotes{
		Start: start,
		Epoch: epoch,
		Votes: records,
	})
}

func (em *Emitter) addLlrEpochVote(e *inter.MutableEventPayload) {
	if em.skipLlrEpochVote() || e.Version() == 0 {
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
	if !em.isEpochValidator(target) {
		return
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

func (em *Emitter) neverSkipLlrVote() bool {
	return em.stakeRatio[em.config.Validator.ID] <= uint64(piecefunc.DecimalUnit)/3+1
}

func (em *Emitter) skipLlrBlockVote() bool {
	if em.neverSkipLlrVote() {
		return false
	}
	// poor validators vote only if we have a long batch of non-decided blocks
	return em.world.GetLatestBlockIndex() < em.world.GetLowestBlockToDecide()+basiccheck.MaxBlockVotesPerEvent*3
}

func (em *Emitter) skipLlrEpochVote() bool {
	if em.neverSkipLlrVote() {
		return false
	}
	// poor validators vote if we have a long batch of non-decided epochs
	if em.epoch > em.world.GetLowestEpochToDecide()+2 {
		return false
	}
	// otherwise, poor validators have a small chance to vote
	return rand.Intn(30) != 0
}

func (em *Emitter) isEpochValidator(epoch idx.Epoch) bool {
	es := em.world.GetHistoryEpochState(epoch)
	if es == nil {
		return false
	}

	_, ok := es.ValidatorProfiles[em.config.Validator.ID]
	return ok
}
