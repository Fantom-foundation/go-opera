package posposet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPosetRush(t *testing.T) {
	assert := assert.New(t)

	nodes, eventsByNode := GenEventsByNode(6, 90, 3)
	p := FakePoset(nodes)

	t.Run("Multiple start", func(t *testing.T) {
		p.Stop()
		p.Start()
		p.Start()
	})

	t.Run("Unordered event stream", func(t *testing.T) {
		// push events in reverse order
		for _, events := range eventsByNode {
			for i := len(events) - 1; i >= 0; i-- {
				e := events[i]
				p.PushEventSync(*e)
			}
		}
		// check all events are in poset store
		for _, events := range eventsByNode {
			for _, e0 := range events {
				e1 := p.store.GetEvent(e0.Hash())
				if !assert.NotNil(e1, "Event is not in poset store") {
					return
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
 * Poset's test methods:
 */

// PushEventSync takes event into processing. It's a sync version of Poset.PushEvent().
// Event order doesn't matter.
func (p *Poset) PushEventSync(e Event) {
	initEventIdx(&e)

	p.onNewEvent(&e)
}
