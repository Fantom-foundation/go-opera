package posposet

import (
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

func TestPoset(t *testing.T) {
	nodes, eventsByNode := GenEventsByNode(4, 10, 3)
	p := FakePoset(nodes)

	t.Run("Multiple start", func(t *testing.T) {
		p.Stop()
		p.Start()
		p.Start()
	})

	t.Run("Push unordered events", func(t *testing.T) {
		// push events in reverse order
		for _, events := range eventsByNode {
			for i := len(events) - 1; i >= 0; i-- {
				e := events[i]
				err := p.PushEventSync(*e)
				if err != nil {
					t.Fatal(err)
				}
			}
		}

		// check all events are in poset store
		for _, events := range eventsByNode {
			for _, e0 := range events {
				e1 := p.store.GetEvent(e0.Hash())
				if e1 == nil {
					t.Fatal("Event is not in poset store")
				}
			}
		}
	})

	t.Run("Multiple stop", func(t *testing.T) {
		p.Stop()
		p.Stop()
	})
}

/*
 * Utils:
 */

// FakePoset creates
func FakePoset(nodes []common.Address) *Poset {
	balances := make(map[common.Address]uint64, len(nodes))
	for _, addr := range nodes {
		balances[addr] = uint64(10)
	}

	store := NewMemStore()
	err := store.ApplyGenesis(balances)
	if err != nil {
		panic(err)
	}

	p := New(store)
	return p
}

// GenEventsByNode generates random events.
// Result:
//   - nodes  is an array of node addresses;
//   - events maps node address to array of its events;
func GenEventsByNode(nodeCount, eventCount, parentCount int) (
	nodes []common.Address, events map[common.Address][]*Event) {
	// init results
	nodes = make([]common.Address, nodeCount)
	events = make(map[common.Address][]*Event, nodeCount)
	// make nodes
	for i := 0; i < nodeCount; i++ {
		nodes[i] = common.FakeAddress()
	}
	// make events
	for i := 0; i < nodeCount*eventCount; i++ {
		// make event with random parents
		parents := rand.Perm(nodeCount)
		creator := nodes[parents[0]]
		e := &Event{
			Creator: creator,
			Parents: EventHashes{},
		}
		// first parent is a last creator's event or empty hash
		if ee := events[creator]; len(ee) > 0 {
			e.Parents.Add(ee[len(ee)-1].Hash())
		} else {
			e.Parents.Add(EventHash{})
		}
		// other parents are the lasts other's events
		others := parentCount
		for _, other := range parents[1:] {
			if others--; others < 0 {
				break
			}
			if ee := events[nodes[other]]; len(ee) > 0 {
				e.Parents.Add(ee[len(ee)-1].Hash())
			}
		}
		// save event
		events[creator] = append(events[creator], e)
	}

	return
}

/*
 * Poset test methods:
 */

// PushEventSync takes event into processing. It's a sync version of Poset.PushEvent().
// Event order doesn't matter.
func (p *Poset) PushEventSync(e Event) error {
	err := initEventIdx(&e)
	if err != nil {
		return err
	}
	p.onNewEvent(&e)
	return nil
}
