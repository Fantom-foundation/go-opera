package posposet

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPosetRush(t *testing.T) {
	return // NOTE: temporary
	nodes, eventsByNode := GenEventsByNode(4, 10, 3)
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

func TestPosetRoots(t *testing.T) {
	assert := assert.New(t)
	// node name means:
	// - 1st letter uppercase - node should be root;
	// - 2nd number - index by node;
	// - 3rd number - frame where node should be in;
	nodes, _, names := ParseEvents(`
a01   b01   c01   
║     ║     ║     
a11 ─ ╬ ─ ─ ╣     d01
║     ║     ║     ║
║     ╠ ─ ─ c11 ─ ╣
║     ║     ║     ║     e01
╠ ─ ─ B12 ─ ╣     ║     ║
║     ║     ║     ║     ║
║     ║     ╠ ─ ─ D12 ─ ╣
║     ║     ║     ║     ║
A22 ─ ╫ ─ ─ ╬ ─ ─ ╣     ║
║     ║     ║     ║     ║
╠ ─ ─ ╫ ─ ─ ╫ ─ ─ ╬ ─ ─ E12
║     ║     ║     ║     ║
╠ ─ ─ ╫ ─ ─ C22 ─ ╣     ║
║     ║     ║     ║     ║
╠ ─ ─ B23 ─ ╣     ║     ║
║     ║     ║     ║     ║
║     ║     ╠ ─ ─ D23 ─ ╣
║     ║     ║     ║     ║
║     ╠ ─ ─ ╫ ─ ─ ╬ ─ ─ E23
║     ║     ║     ║     ║
A33 ─ ╬ ─ ─ ╣     ║     ║
║     ║     ║     ║     ║
║     ╠ ─ ─ C33   ║     ║
║     ║     ║     ║     ║
╠ ─ ─ b33 ─ ╣     ║     ║
║     ║     ║     ║     ║
a43 ─ ╬ ─ ─ ╣     ║     ║
║     ║     ║     ║     ║
║     ╠ ─ ─ C44 ─ ╣     ║
║     ║     ║     ║     ║
╠ ─ ─ B44 ─ ╣     ║     ║
║     ║     ║     ║     ║
║     ║     ╠ ─ ─ D34 ─ ╣
║     ║     ║     ║     ║
A54 ─ ╫ ─ ─ ╬ ─ ─ ╣     ║
║     ║     ║     ║     ║
╠ ─ ─ ╫ ─ ─ c54 ─ ╣     ║
║     ║     ║     ║     ║
║     ║     ╠ ─ ─ ╬ ─ ─ E34
║     ║     ║     ║     ║
`)
	p := FakePoset(nodes)
	// process events
	for _, e := range names {
		p.PushEventSync(*e)
	}

	for name, e := range names {
		// check roots
		mustBeRoot := (name == strings.ToUpper(name))
		isReallyRoot := p.RootFrame(e) != nil
		if !assert.Equal(mustBeRoot, isReallyRoot, name) {
			break
		}
		if !isReallyRoot {
			continue
		}
		// check frames
		mustBeFrame, err := strconv.ParseUint(name[2:3], 10, 64)
		if !assert.NoError(err, "name the nodes properly: <UpperCaseForRoot><Index><FrameN>") {
			return
		}
		reallyFrame := p.RootFrame(e)
		if !assert.Equal(mustBeFrame, *reallyFrame, name) {
			break
		}
	}
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
