package posposet

import (
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/ordering"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

var (
	genesisTestTime = inter.Timestamp(1565000000 * time.Second)
)

// FakePoset creates empty poset with mem store and equal stakes of nodes in genesis.
// Input event order doesn't matter.
func FakePoset(nodes []hash.Peer) (*Poset, *Store, *EventStore) {
	balances := make(map[hash.Peer]inter.Stake, len(nodes))
	for _, addr := range nodes {
		balances[addr] = inter.Stake(1)
	}

	store := NewMemStore()
	err := store.ApplyGenesis(balances, genesisTestTime)
	if err != nil {
		panic(err)
	}

	input := NewEventStore(nil)

	poset := New(store, input)
	poset.Bootstrap()
	MakeOrderedInput(poset)
	poset.Start()

	return poset, store, input
}

// MakeOrderedInput wraps Poset.onNewEvent with ordering.EventBuffer.
// For tests only.
func MakeOrderedInput(p *Poset) {
	processed := make(hash.EventsSet) // NOTE: mem leak, so for tests only.

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

// ASCIIschemeToDAG wrap inter.ASCIIschemeToDAG() to prepare events properly.
func ASCIIschemeToDAG(
	scheme string,
	mods ...func(e *inter.Event, name string) *inter.Event,
) (
	nodes []hash.Peer,
	events map[hash.Peer][]*inter.Event,
	names map[string]*inter.Event,
) {
	// get nodes only
	nodes, _, _ = inter.ASCIIschemeToDAG(scheme)
	// init poset
	p, _, input := FakePoset(nodes)

	buildEvent := func(e *inter.Event, name string) *inter.Event {
		e.Epoch = p.CurrentSuperFrameN()
		e = p.Prepare(e)
		e.RecacheHash()

		input.SetEvent(e)
		p.PushEventSync(e.Hash())

		return e
	}

	mods = append(mods, buildEvent)

	// process events
	nodes, events, names = inter.ASCIIschemeToDAG(scheme, mods...)

	return
}
