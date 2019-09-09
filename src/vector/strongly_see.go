package vector

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// StronglySee calculates "sufficient coherence" between the events.
// The A.HighestBefore array remembers the sequence number of the last
// event by each member that is an ancestor of A. The array for
// B.LowestAfter remembers the sequence number of the earliest
// event by each member that is a descendant of B. Compare the two arrays,
// and find how many elements in the A.HighestBefore array are greater
// than or equal to the corresponding element of the B.LowestAfter
// array. If there are more than 2n/3 such matches, then the A and B
// have achieved sufficient coherency.
func (vi *Index) StronglySee(aID, bID hash.Event) bool {
	// get events by hash
	a := vi.GetHighestBeforeSeq(aID)
	if a == nil {
		vi.Log.Error("Event A wasn't found", "event", aID.String())
		return false
	}
	b := vi.GetLowestAfterSeq(bID)
	if b == nil {
		vi.Log.Error("Event B wasn't found", "event", bID.String())
		return false
	}

	yes := vi.members.NewCounter()
	no := vi.members.NewCounter()

	// calculate strongly seeing using the indexes
	for creator, n := range vi.memberIdxs {
		bLowestAfter := b.Get(n)
		aHighestBefore := a.Get(n).Seq

		if bLowestAfter <= aHighestBefore && bLowestAfter != 0 {
			yes.Count(creator)
		} else {
			no.Count(creator)
		}

		if yes.HasQuorum() {
			return true
		}

		if no.HasQuorum() {
			return false
		}
	}

	return false
}
