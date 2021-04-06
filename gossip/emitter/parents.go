package emitter

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/emitter/ancestor"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

// buildSearchStrategies returns a strategy for each parent search
func (em *Emitter) buildSearchStrategies(maxParents idx.Event) []ancestor.SearchStrategy {
	strategies := make([]ancestor.SearchStrategy, 0, maxParents)
	if maxParents == 0 {
		return strategies
	}
	payloadStrategy := em.payloadIndexer.SearchStrategy()
	for idx.Event(len(strategies)) < 1 {
		strategies = append(strategies, payloadStrategy)
	}
	randStrategy := ancestor.NewRandomStrategy(nil)
	for idx.Event(len(strategies)) < maxParents/2 {
		strategies = append(strategies, randStrategy)
	}
	quorumStrategy := em.quorumIndexer.SearchStrategy()
	for idx.Event(len(strategies)) < maxParents {
		strategies = append(strategies, quorumStrategy)
	}
	return strategies
}

// chooseParents selects an "optimal" parents set for the validator
func (em *Emitter) chooseParents(epoch idx.Epoch, myValidatorID idx.ValidatorID) (*hash.Event, hash.Events, bool) {
	selfParent := em.world.GetLastEvent(epoch, myValidatorID)
	heads := em.world.GetHeads(epoch) // events with no descendants

	if selfParent != nil && len(em.world.DagIndex().NoCheaters(selfParent, hash.Events{*selfParent})) == 0 {
		em.Periodic.Error(time.Second, "Events emitting isn't allowed due to the doublesign", "validator", myValidatorID)
		return nil, nil, false
	}

	var parents hash.Events
	if selfParent != nil {
		parents = hash.Events{*selfParent}
	}
	parents = ancestor.ChooseParents(parents, heads, em.buildSearchStrategies(em.maxParents-idx.Event(len(parents))))
	return selfParent, parents, true
}
