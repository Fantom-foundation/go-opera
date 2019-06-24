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
	balances := make(map[hash.Peer]uint64, len(nodes))
	for _, addr := range nodes {
		balances[addr] = uint64(1)
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
func MakeOrderedInput(p *Poset) {
	orderThenConsensus := ordering.EventBuffer(
		// process
		p.consensus,
		// drop
		func(e *inter.Event, err error) {
			logger.Get().Warn(err.Error() + ", so rejected")
		},
		// exists
		func(h hash.Event) *inter.Event {
			if p.store.GetEventFrame(h) == nil {
				return nil
			}
			return p.input.GetEvent(h)
		},
	)
	// event order doesn't matter now
	p.onNewEvent = func(e *inter.Event) {
		orderThenConsensus(e)
	}
}
