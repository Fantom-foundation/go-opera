package vecmt

import (
	"errors"

	"github.com/Fantom-foundation/lachesis-base/hash"
)

// NoCheaters excludes events which are observed by selfParents as cheaters.
// Called by emitter to exclude cheater's events from potential parents list.
func (vi *Index) NoCheaters(selfParent *hash.Event, options hash.Events) hash.Events {
	if selfParent == nil {
		return options
	}
	vi.InitBranchesInfo()

	if !vi.Engine.AtLeastOneFork() {
		return options
	}

	// no need to merge, because every branch is marked by IsForkDetected if fork is observed
	highest := vi.Base.GetHighestBefore(*selfParent)
	filtered := make(hash.Events, 0, len(options))
	for _, id := range options {
		e := vi.getEvent(id)
		if e == nil {
			vi.crit(errors.New("event not found"))
		}
		if !highest.Get(vi.validatorIdxs[e.Creator()]).IsForkDetected() {
			filtered.Add(id)
		}
	}
	return filtered
}
