package election

import (
	"errors"
	"sort"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
)

type sortedRoot struct {
	stake pos.Stake
	root  hash.Event
}
type sortedRoots []sortedRoot

func (s sortedRoots) Len() int {
	return len(s)
}
func (s sortedRoots) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// compare by stake amount, root hash
func (s sortedRoots) Less(i, j int) bool {
	if s[i].stake != s[j].stake {
		return s[i].stake > s[j].stake
	}
	return s[i].root.Big().Cmp(s[j].root.Big()) > 0
}

// Chooses the decided "yes" roots with the greatest stake amount.
// This root serves as a "checkpoint" within DAG, as it's guaranteed to be final and consistent unless more than 1/3n are Byzantine.
// Other validators will come to the same Atropos not later than current highest frame + 2.
func (el *Election) chooseAtropos() (*ElectionRes, error) {
	finalRoots := make(sortedRoots, 0, len(el.validators))
	// fill yesRoots
	for validator, stake := range el.validators {
		vote, ok := el.decidedRoots[validator]
		if !ok {
			el.Log.Crit("Called before all the roots are decided")
		}
		if vote.yes {
			finalRoots = append(finalRoots, sortedRoot{
				root:  vote.observedRoot,
				stake: stake,
			})
		}
	}
	if len(finalRoots) == 0 {
		return nil, errors.New("All the roots aren't decided as 'yes', which is possible only if more than 1/3n are Byzantine")
	}

	// sort by stake amount, root hash
	sort.Sort(finalRoots)

	// take root with greatest stake
	return &ElectionRes{
		Frame:   el.frameToDecide,
		Atropos: finalRoots[0].root,
	}, nil
}
