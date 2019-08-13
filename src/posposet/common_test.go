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

type BufferedPoset struct {
	*Poset

	bufferPush ordering.PushEventFn
}

func (p *BufferedPoset) PushToBuffer(e *inter.Event) {
	p.bufferPush(e, "")
}

// FakePoset creates empty poset with mem store and equal stakes of nodes in genesis.
// Input event order doesn't matter.
func FakePoset(nodes []hash.Peer) (*BufferedPoset, *Store, *EventStore) {
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

	buffered := &BufferedPoset{
		Poset: poset,
		bufferPush: MakeOrderedInput(poset),
	}

	return buffered, store, input
}

// MakeOrderedInput wraps Poset.onNewEvent with ordering.EventBuffer.
// For tests only.
func MakeOrderedInput(p *Poset) ordering.PushEventFn {
	processed := make(hash.EventsSet) // NOTE: mem leak, so for tests only.

	orderThenConsensus, _ := ordering.EventBuffer(ordering.Callback{

		Process: func(event *inter.Event) error {
			processed.Add(event.Hash())
			return p.ProcessEvent(event)
		},

		Drop: func(e *inter.Event, peer string, err error) {
			logger.Get().Warn(err.Error() + ", so rejected")
		},

		Exists: func(h hash.Event) *inter.Event {
			if _, ok := processed[h]; ok {
				return p.input.GetEvent(h)
			}
			return nil
		},
	})
	return orderThenConsensus
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
		err := p.ProcessEvent(e)
		if err != nil {
			p.Fatal(err)
		}

		return e
	}

	mods = append(mods, buildEvent)

	// process events
	nodes, events, names = inter.ASCIIschemeToDAG(scheme, mods...)

	return
}
