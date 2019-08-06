package ordering

import (
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

func TestEventBuffer(t *testing.T) {
	nodes := inter.GenNodes(5)
	events := inter.GenEventsByNode(nodes, 10, 3, nil, nil)
	processed := make(map[hash.Event]*inter.Event)

	push := EventBuffer(Callback{

		Process: func(e *inter.Event) {
			if _, ok := processed[e.Hash()]; ok {
				t.Fatalf("%s already processed", e.String())
				return
			}
			for _, p := range e.Parents {
				if _, ok := processed[p]; !ok {
					t.Fatalf("got %s before parent %s", e.String(), p.String())
					return
				}
			}
			processed[e.Hash()] = e
		},

		Drop: func(e *inter.Event, err error) {
			t.Fatalf("%s unexpectedly dropped with %s", e.String(), err)
		},

		Exists: func(e hash.Event) *inter.Event {
			return processed[e]
		},
	})

	for _, ee := range events {
		for _, e := range ee {
			push(e)
		}
	}
}
