package vectorindex

import (
	"sort"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// Handy index for the MedianTime() func
type medianTimeIndex struct {
	stake       inter.Stake
	claimedTime inter.Timestamp
}

// MedianTime calculates weighted median of claimed time within highest seen events.
func (vi *Vindex) MedianTime(id hash.Event) inter.Timestamp {
	// get event by hash
	event, ok := vi.events[id]
	if !ok {
		vi.Error("Vindex: event wasn't found " + id.String())
		return 0
	}

	// convert []HighestBefore -> []medianTimeIndex
	highests := make([]medianTimeIndex, 0, len(event.HighestBefore))
	for creator, n := range vi.memberIdxs {
		highests = append(highests, medianTimeIndex{
			stake:       vi.members[creator],
			claimedTime: event.HighestBefore[n].ClaimedTime,
		})
	}

	// sort by claimed time (partial order is enough here, because we need only claimedTime)
	sort.Slice(highests, func(i, j int) bool {
		a, b := highests[i], highests[j]
		return a.claimedTime < b.claimedTime
	})

	// Calculate weighted median
	totalStake := vi.members.TotalStake()
	halfStake := totalStake / 2
	var currStake inter.Stake
	var median inter.Timestamp
	for _, highest := range highests {
		currStake += highest.stake
		if currStake >= halfStake {
			median = highest.claimedTime
			break
		}
	}

	// sanity check
	if currStake < halfStake || currStake > totalStake {
		vi.Fatalf("Vindex: median wasn't calculated correctly (median=%d, currStake=%d, totalStake=%d, len(highests)=%d, id=%s)", median, currStake, totalStake, len(highests), id.String())
	}

	return median
}
