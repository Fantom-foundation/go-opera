package posposet

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

func TestPoset(t *testing.T) {
	nodes, nodesEvents := GenEventsByNode(5, 99, 3)

	posets := make([]*Poset, len(nodes))
	inputs := make([]*EventStore, len(nodes))
	for i := 0; i < len(nodes); i++ {
		posets[i], _, inputs[i] = FakePoset(nodes)
		posets[i].SetName(nodes[i].String())
		posets[i].store.SetName(nodes[i].String())
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
				inputs[n].SetEvent(e.Event)
				posets[n].PushEventSync(e.Hash())
			}
		}
		// second all events from others
		for n := 0; n < len(nodes); n++ {
			events := nodesEvents[nodes[n]]
			for _, e := range events {
				for i := 0; i < len(posets); i++ {
					if i != n {
						inputs[i].SetEvent(e.Event)
						posets[i].PushEventSync(e.Hash())
					}
				}
			}
		}
	})

	t.Run("All events in Store", func(t *testing.T) {
		assertar := assert.New(t)
		for _, events := range nodesEvents {
			for _, e0 := range events {
				frame := posets[0].store.GetEventFrame(e0.Hash())
				if !assertar.NotNil(frame, "Event is not in poset store") {
					return
				}
			}
		}
	})

	t.Run("Check consensus", func(t *testing.T) {
		assertar := assert.New(t)
		for i := 0; i < len(posets)-1; i++ {
			for j := i + 1; j < len(posets); j++ {
				p0, p1 := posets[i], posets[j]
				// compare blockchain
				if !assertar.Equal(p0.state.LastBlockN, p1.state.LastBlockN, "blocks count") {
					s0 := getASCIIByPoset(t, p0, nodesEvents, len(nodes))
					s1 := getASCIIByPoset(t, p1, nodesEvents, len(nodes))

					out := joinStrings(s0, s1)

					t.Log(out)

					return
				}

				for b := uint64(1); b <= p0.state.LastBlockN; b++ {
					if !assertar.Equal(p0.store.GetBlock(b), p1.store.GetBlock(b), "block") {
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

// PushEventSync takes event into processing.
// It's a sync version of Poset.PushEvent().
func (p *Poset) PushEventSync(e hash.Event) {
	event := p.input.GetEvent(e)
	p.onNewEvent(event)
}

func getASCIIByPoset(t *testing.T, poset *Poset, nodesEvents map[hash.Peer][]*Event, nodeCount int) string {
	events := inter.Events{}
	for _, ee := range nodesEvents {
		for _, e := range ee {
			e = poset.GetEvent(e.Hash())
			events = append(events, e.Event)
		}
	}

	scheme, err := inter.DAGtoASCIIcheme(events, nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	return scheme
}

// join side-by-side
func joinStrings(str0, str1 string) string {
	var ar0, ar1 []string

	scanner := bufio.NewScanner(strings.NewReader(str0))
	for scanner.Scan() {
		ar0 = append(ar0, scanner.Text())
	}

	scanner = bufio.NewScanner(strings.NewReader(str1))
	for scanner.Scan() {
		ar1 = append(ar1, scanner.Text())
	}

	var res strings.Builder
	for i, str := range ar0 {
		res.WriteString("\n" + str + "\t\t" + ar1[i])
	}

	return res.String()
}
