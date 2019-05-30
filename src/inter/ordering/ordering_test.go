package ordering

import (
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

func TestEventBuffer(t *testing.T) {
	_, events := inter.GenEventsByNode(5, 10, 3)
	processed := make(map[hash.Event]*inter.Event)

	process := func(e *inter.Event) {
		if _, ok := processed[e.Hash()]; ok {
			t.Fatalf("%s already processed", e.String())
			return
		}
		for _, p := range e.Parents.Slice() {
			if p.IsZero() {
				continue
			}
			if _, ok := processed[p]; !ok {
				t.Fatalf("got %s before parent %s", e.String(), p.String())
				return
			}
		}
		processed[e.Hash()] = e
	}

	exists := func(e hash.Event) *inter.Event {
		return processed[e]
	}

	drop := func(e *inter.Event, err error) {
		t.Fatalf("%s unexpectedly dropped with %s", e.String(), err)
	}

	push := EventBuffer(process, drop, exists)
	for _, ee := range events {
		for _, e := range ee {
			push(e)
		}
	}
}
