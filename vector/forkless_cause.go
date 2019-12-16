package vector

import (
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

type kv struct {
	a, b hash.Event
}

// ForklessCause calculates "sufficient coherence" between the events.
// The A.HighestBefore array remembers the sequence number of the last
// event by each validator that is an ancestor of A. The array for
// B.LowestAfter remembers the sequence number of the earliest
// event by each validator that is a descendant of B. Compare the two arrays,
// and find how many elements in the A.HighestBefore array are greater
// than or equal to the corresponding element of the B.LowestAfter
// array. If there are more than 2n/3 such matches, then the A and B
// have achieved sufficient coherency.
//
// If B1 and B2 are forks, then they cannot BOTH forkless-cause any specific event A,
// unless more than 1/3W are Byzantine.
// This great property is the reason why this function exists,
// providing the base for the BFT algorithm.
func (vi *Index) ForklessCause(aID, bID hash.Event) bool {
	if res, ok := vi.cache.ForklessCause.Get(kv{aID, bID}); ok {
		return res.(bool)
	}

	res := vi.forklessCause(aID, bID)

	vi.cache.ForklessCause.Add(kv{aID, bID}, res)
	return res
}

func (vi *Index) forklessCause(aID, bID hash.Event) bool {
	vi.initBranchesInfo()

	// get events by hash
	a := vi.GetHighestBeforeSeq(aID)
	if a == nil {
		vi.Log.Crit("Event A not found", "event", aID.String())
		return false
	}

	// check A doesn't observe any forks from B
	if vi.atLeastOneFork() {
		bBranchID := vi.getEventBranchID(bID)
		if a.Get(bBranchID).IsForkDetected() { // B is observed as cheater by A
			return false
		}
	}

	// check A observes that {QUORUM} non-cheater-validators observe B
	b := vi.GetLowestAfterSeq(bID)
	if b == nil {
		vi.Log.Crit("Event B not found", "event", bID.String())
		return false
	}

	yes := vi.validators.NewCounter()
	// calculate forkless causing using the indexes
	for branchIDint, creatorIdx := range vi.bi.BranchIDCreatorIdxs {
		branchID := idx.Validator(branchIDint)

		// bLowestAfter := vi.GetLowestAfterSeq_(bID, branchID)   // lowest event from creator on branchID, which observes B
		bLowestAfter := b.Get(branchID)   // lowest event from creator on branchID, which observes B
		aHighestBefore := a.Get(branchID) // highest event from creator, observed by A

		// if lowest event from branchID which observes B <= highest from branchID observed by A
		// then {highest from branchID observed by A} observes B
		if bLowestAfter <= aHighestBefore.Seq && bLowestAfter != 0 && !aHighestBefore.IsForkDetected() {
			// we may count the same creator multiple times (on different branches)!
			// so not every call increases the counter
			yes.CountByIdx(creatorIdx)
		}
	}
	return yes.HasQuorum()
}

// NoCheaters excludes events which are observed by selfParents as cheaters.
// Called by emitter to exclude cheater's events from potential parents list.
func (vi *Index) NoCheaters(selfParent *hash.Event, options hash.Events) hash.Events {
	if selfParent == nil {
		return options
	}
	vi.initBranchesInfo()

	// no need to merge, because every branch is marked by IsForkDetected if fork is observed
	highest := vi.GetHighestBeforeSeq(*selfParent)
	filtered := make(hash.Events, 0, len(options))
	for _, id := range options {
		header := vi.getEvent(id)
		if header == nil {
			vi.Log.Crit("Event not found", "id", id.String())
		}
		if !highest.Get(vi.validatorIdxs[header.Creator]).IsForkDetected() {
			filtered.Add(id)
		}
	}
	return filtered
}
