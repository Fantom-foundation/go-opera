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
func (vi *Vindex) MedianTime(id hash.Event, genesisTime inter.Timestamp) inter.Timestamp {
	// get event by hash
	event := vi.GetEvent(id)
	if event == nil {
		vi.Error("Vindex: event wasn't found " + id.String())
		return 0
	}

	honestTotalStake := inter.Stake(0) // isn't equal to members.TotalStake(), because doesn't count cheaters
	highests := make([]medianTimeIndex, 0, len(event.HighestBefore))
	// convert []HighestBefore -> []medianTimeIndex
	for creator, n := range vi.memberIdxs {
		highest := medianTimeIndex{}
		highest.stake = vi.members[creator]
		highest.claimedTime = event.HighestBefore[n].ClaimedTime

		// edge cases
		if event.HighestBefore[n].IsForkSeen {
			// cheaters don't influence medianTime
			highest.stake = 0
		} else if event.HighestBefore[n].Seq == 0 {
			// if no event was seen from this node, then use genesisTime
			highest.claimedTime = genesisTime
		}

		highests = append(highests, highest)
		honestTotalStake += highest.stake
	}
	// it's technically possible totalStake == 0 (all members are cheaters)

	// sort by claimed time (partial order is enough here, because we need only claimedTime)
	sort.Slice(highests, func(i, j int) bool {
		a, b := highests[i], highests[j]
		return a.claimedTime < b.claimedTime
	})

	// Calculate weighted median
	halfStake := honestTotalStake / 2
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
	if currStake < halfStake || currStake > honestTotalStake {
		vi.Fatalf("Vindex: median wasn't calculated correctly (median=%d, currStake=%d, totalStake=%d, len(highests)=%d, id=%s)", median, currStake, honestTotalStake, len(highests), id.String())
	}

	return median
}
