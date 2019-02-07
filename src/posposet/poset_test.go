package posposet

import (
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

func TestPoset(t *testing.T) {
	p := FakePoset()

	t.Run("Multiple start", func(t *testing.T) {
		p.Stop()
		p.Start()
		p.Start()
	})

	t.Run("Push unordered events", func(t *testing.T) {
		eventsByNode := GenEventsByNode(4, 10, 3)
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

func FakePoset() *Poset {
	store := NewMemStore()
	p := New(store)
	return p
}

func GenEventsByNode(nodes, events, maxParents int) map[common.Address][]*Event {
	res := make(map[common.Address][]*Event, nodes)

	// make nodes
	nn := make([]common.Address, nodes)
	for i := 0; i < nodes; i++ {
		node := common.FakeAddress()
		nn[i] = node
		res[node] = nil
	}

	for i := 0; i < nodes*events; i++ {
		// make event with random parents
		parents := rand.Perm(nodes)
		creator := nn[parents[0]]
		e := &Event{
			Creator: creator,
			Parents: EventHashes{},
		}
		// first parent is a last creator's event or empty hash
		if ee := res[creator]; len(ee) > 0 {
			e.Parents = append(e.Parents, ee[len(ee)-1].Hash())
		} else {
			e.Parents = append(e.Parents, EventHash{})
		}
		// other parents are the lasts other's events
		others := maxParents
		for _, other := range parents[1:] {
			if others--; others < 0 {
				break
			}
			if ee := res[nn[other]]; len(ee) > 0 {
				e.Parents = append(e.Parents, ee[len(ee)-1].Hash())
			}
		}
		// save event
		res[creator] = append(res[creator], e)
	}

	return res
}

/*
 * Poset methods:
 */

// PushEventSync takes event into processing. Event order doesn't matter.
func (p *Poset) PushEventSync(e Event) error {
	err := initEventIdx(&e)
	if err != nil {
		return err
	}
	p.onNewEvent(&e)
	return nil
}
