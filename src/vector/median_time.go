package vector

import (
	"sort"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

// Handy index for the MedianTime() func
type medianTimeIndex struct {
	stake       pos.Stake
	claimedTime inter.Timestamp
}

// MedianTime calculates weighted median of claimed time within highest observed events.
func (vi *Index) MedianTime(id hash.Event, genesisTime inter.Timestamp) inter.Timestamp {
	// get event by hash
	observed := vi.GetHighestBeforeSeq(id)
	times := vi.GetHighestBeforeTime(id)
	if observed == nil || times == nil {
		vi.Log.Error("Event wasn't found", "event", id.String())

		return 0
	}

	honestTotalStake := pos.Stake(0) // isn't equal to validators.TotalStake(), because doesn't count cheaters
	highests := make([]medianTimeIndex, 0, len(vi.validatorIdxs))
	// convert []HighestBefore -> []medianTimeIndex
	for creator, n := range vi.validatorIdxs {
		highest := medianTimeIndex{}
		highest.stake = vi.validators[creator]
		highest.claimedTime = times.Get(n)

		// edge cases
		forkSeq := observed.Get(n)
		if forkSeq.IsForkDetected {
			// cheaters don't influence medianTime
			highest.stake = 0
		} else if forkSeq.Seq == 0 {
			// if no event was observed from this node, then use genesisTime
			highest.claimedTime = genesisTime
		}

		highests = append(highests, highest)
		honestTotalStake += highest.stake
	}
	// it's technically possible totalStake == 0 (all validators are cheaters)

	// sort by claimed time (partial order is enough here, because we need only claimedTime)
	sort.Slice(highests, func(i, j int) bool {
		a, b := highests[i], highests[j]
		return a.claimedTime < b.claimedTime
	})

	// Calculate weighted median
	halfStake := honestTotalStake / 2
	var currStake pos.Stake
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
		vi.Log.Crit("Median wasn't calculated correctly",
			"median", median,
			"currStake", currStake,
			"totalStake", honestTotalStake,
			"len(highests)", len(highests),
			"id", id.String(),
		)
	}

	return median
}
