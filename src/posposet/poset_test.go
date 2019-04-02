package posposet

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

func TestPoset(t *testing.T) {
	nodes, nodesEvents := GenEventsByNode(5, 99, 3)

	posets := make([]*Poset, len(nodes))
	inputs := make([]*EventStore, len(nodes))
	for i := 0; i < len(nodes); i++ {
		posets[i], _, inputs[i] = FakePoset(nodes)
	}

	t.Run("Multiple start", func(t *testing.T) {
		posets[0].Stop()
		posets[0].Start()
		posets[0].Start()
	})

	t.Run("Push unordered events", func(t *testing.T) {
		// first all events from one node
		for n := 0; n < len(nodes); n++ {
			events := nodesEvents[nodes[n]]
			for _, e := range events {
				inputs[n].SetEvent(&e.Event)
				posets[n].PushEventSync(e.Hash())
			}
		}
		// second all events from others
		for n := 0; n < len(nodes); n++ {
			events := nodesEvents[nodes[n]]
			for _, e := range events {
				for i := 0; i < len(posets); i++ {
					if i != n {
						inputs[i].SetEvent(&e.Event)
						posets[i].PushEventSync(e.Hash())
					}
				}
			}
		}
	})

	t.Run("All events in Store", func(t *testing.T) {
		assert := assert.New(t)
		for _, events := range nodesEvents {
			for _, e0 := range events {
				frame := posets[0].store.GetEventFrame(e0.Hash())
				if !assert.NotNil(frame, "Event is not in poset store") {
					return
				}
			}
		}
	})

	t.Run("Check consensus", func(t *testing.T) {
		assert := assert.New(t)
		for i := 0; i < len(posets)-1; i++ {
			for j := i + 1; j < len(posets); j++ {
				p0, p1 := posets[i], posets[j]
				// compare blockchain
				if !assert.Equal(p0.state.LastBlockN, p1.state.LastBlockN, "blocks count") {
					return
				}
				for b := uint64(1); b <= p0.state.LastBlockN; b++ {
					if !assert.Equal(p0.store.GetBlock(b), p1.store.GetBlock(b), "block") {
						return
					}
				}
			}
		}
	})

	t.Run("Multiple stop", func(t *testing.T) {
		posets[0].Stop()
		posets[0].Stop()
	})
}

/*
 * Poset's test methods:
 */

// PushEventSync takes event into processing. It's a sync version of Poset.PushEvent().
// Event order doesn't matter.
func (p *Poset) PushEventSync(e hash.Event) {
	event := p.GetEvent(e)
	p.onNewEvent(event)
}
