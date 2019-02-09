package posposet

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

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
				p.PushEventSync(*e)
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

func TestRoots(t *testing.T) {
	assert := assert.New(t)
	nodes, _, names := ParseEvents(`
a00   b00   c00   
║     ║     ║     
a01 ─ ╬ ─ ─ ╣     d00
║     ║     ║     ║
║     ╠ ─ ─ c01 ─ ╣
║     ║     ║     ║     e00
╠ ─ ─ B01 ─ ╣     ║     ║
║     ║     ║     ║     ║
║     ║     ╠ ─ ─ D01 ─ ╣
║     ║     ║     ║     ║
A02 ─ ╫ ─ ─ ╬ ─ ─ ╣     ║
║     ║     ║     ║     ║
`)
	p := FakePoset(nodes)
	// process events
	for _, e := range names {
		p.PushEventSync(*e)
	}
	// check roots
	for name, e := range names {
		mustBeRoot := (name == strings.ToUpper(name))
		isReallyRoot := p.frame(p.state.LastFinishedFrameN + 1).IsRoot(e.Hash())
		assert.Equal(mustBeRoot, isReallyRoot, name)
	}
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

/*
 * Poset test methods:
 */

// PushEventSync takes event into processing. It's a sync version of Poset.PushEvent().
// Event order doesn't matter.
func (p *Poset) PushEventSync(e Event) {
	initEventIdx(&e)

	p.onNewEvent(&e)
}
