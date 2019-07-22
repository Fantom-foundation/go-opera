package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/ordering"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

// FakePoset creates empty poset with mem store and equal stakes of nodes in genesis.
// Input event order doesn't matter.
func FakePoset(nodes []hash.Peer) (*Poset, *Store, *EventStore) {
	balances := make(map[hash.Peer]inter.Stake, len(nodes))
	for _, addr := range nodes {
		balances[addr] = inter.Stake(1)
	}

	store := NewMemStore()
	err := store.ApplyGenesis(balances)
	if err != nil {
		panic(err)
	}

	input := NewEventStore(nil)

	poset := New(store, input)
	poset.Bootstrap()
	MakeOrderedInput(poset)

	return poset, store, input
}

// MakeOrderedInput wraps Poset.onNewEvent with ordering.EventBuffer.
// For tests only.
func MakeOrderedInput(p *Poset) {
	processed := make(hash.Events) // NOTE: mem leak, so for tests only.

	orderThenConsensus := ordering.EventBuffer(ordering.Callback{

		Process: func(event *inter.Event) {
			p.consensus(event)
			processed.Add(event.Hash())
		},

		Drop: func(e *inter.Event, err error) {
			logger.Get().Warn(err.Error() + ", so rejected")
		},

		Exists: func(h hash.Event) *inter.Event {
			if _, ok := processed[h]; ok {
				return p.input.GetEvent(h)
			}
			return nil
		},
	})
	// event order doesn't matter now
	p.onNewEvent = func(e *inter.Event) {
		orderThenConsensus(e)
	}
}

// PushEventSync takes event into processing.
// It's a sync version of Poset.PushEvent().
func (p *Poset) PushEventSync(e hash.Event) {
	event := p.input.GetEvent(e)
	p.onNewEvent(event)
}
