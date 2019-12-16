package vector

import (
	"sort"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
)

// medianTimeIndex is a handy index for the MedianTime() func
type medianTimeIndex struct {
	stake       pos.Stake
	claimedTime inter.Timestamp
}

// MedianTime calculates weighted median of claimed time within highest observed events.
func (vi *Index) MedianTime(id hash.Event, genesisTime inter.Timestamp) inter.Timestamp {
	vi.initBranchesInfo()
	// get event by hash
	beforeSeq, times := vi.getHighestBeforeAllBranchesTime(id)
	if beforeSeq == nil || times == nil {
		vi.Log.Error("Event not found", "event", id.String())

		return 0
	}

	honestTotalStake := pos.Stake(0) // isn't equal to validators.TotalStake(), because doesn't count cheaters
	highests := make([]medianTimeIndex, 0, len(vi.validatorIdxs))
	// convert []HighestBefore -> []medianTimeIndex
	for creatorIdxI := range vi.validators.IDs() {
		creatorIdx := idx.Validator(creatorIdxI)
		highest := medianTimeIndex{}
		highest.stake = vi.validators.GetStakeByIdx(creatorIdx)
		highest.claimedTime = times.Get(creatorIdx)
		seq := beforeSeq.Get(creatorIdx)

		// edge cases
		if seq.IsForkDetected() {
			// cheaters don't influence medianTime
			highest.stake = 0
		} else if seq.Seq == 0 {
			// if no event was observed from this node, then use genesisTime
			highest.claimedTime = genesisTime
		}

		highests = append(highests, highest)
		honestTotalStake += highest.stake
	}
	// it's technically possible honestTotalStake == 0 (all validators are cheaters)

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
