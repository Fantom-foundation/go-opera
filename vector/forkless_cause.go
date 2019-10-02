package vector

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
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
// unless more than 1/3n are Byzantine.
// This great property is the reason why this function exists,
// providing the base for the BFT algorithm.
func (vi *Index) ForklessCause(aID, bID hash.Event) bool {
	if res, ok := vi.forklessCauseCache.Get(kv{aID, bID}); ok {
		return res.(bool)
	}

	// get events by hash
	a := vi.GetHighestBeforeSeq(aID)
	if a == nil {
		vi.Log.Error("Event A not found", "event", aID.String())
		return false
	}
	b := vi.GetLowestAfterSeq(bID)
	if b == nil {
		vi.Log.Error("Event B not found", "event", bID.String())
		return false
	}

	yes := vi.validators.NewCounter()
	no := vi.validators.NewCounter()

	res := false
	// calculate forkless seeing using the indexes
	for creator, n := range vi.validatorIdxs {
		bLowestAfter := b.Get(n)
		aHighestBefore := a.Get(n).Seq

		if bLowestAfter <= aHighestBefore && bLowestAfter != 0 {
			yes.Count(creator)
		} else {
			no.Count(creator)
		}

		if yes.HasQuorum() {
			res = true
			break
		}

		if no.HasQuorum() {
			res = false
			break
		}
	}

	//vi.forklessCauseCache.Add(kv{aID, bID}, res)
	return res
}

// Excludes forbidden by consensus parents
// called by emitter to exclude cheater's events from potential parents list
func (vi *Index) NoCheaters(selfParent *hash.Event, options hash.Events) hash.Events {
	if selfParent == nil {
		return options
	}

	highest := vi.GetHighestBeforeSeq(*selfParent)
	filtered := make(hash.Events, 0, len(options))
	for _, id := range options {
		header := vi.getEvent(id)
		if header == nil {
			vi.Log.Crit("Event not found", "id", id.String())
		}
		if !highest.Get(vi.validatorIdxs[header.Creator]).IsForkDetected {
			filtered.Add(id)
		}
	}
	return filtered
}
