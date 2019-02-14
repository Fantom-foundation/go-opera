package posposet

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

func TestParseEvents(t *testing.T) {
	assert := assert.New(t)

	nodes, events, names := ParseEvents(`
a00 b00   c00 d00
║   ║     ║   ║
a01 ║     ║   ║
║   ╠  ─  c01 ║
a02 ╣     ║   ║
║   ║     ║   ║
╠ ─ ╫ ─ ─ c02 ║
║   b01   ║   ║
║   ╠ ─ ─ ╫ ─ d01
║   ║     ║   ║
║   ║     ║   ║
╠ ═ b02 ═ ╬ ═ ╣
║   ║     ║   ║
a03 ╣     ╠ ─ d02
║   ║     ║   ║
║   ║     ║   ╠ ─ e00
║   ║     ║   ║   ║
`)
	expected := map[string][]string{
		"a00": {""},
		"a01": {"a00"},
		"a02": {"a01", "b00"},
		"a03": {"a02", "b02"},
		"b00": {""},
		"b01": {"b00"},
		"b02": {"b01", "a02", "c02", "d01"},
		"c00": {""},
		"c01": {"c00", "b00"},
		"c02": {"c01", "a02"},
		"d00": {""},
		"d01": {"d00", "b01"},
		"d02": {"d01", "c02"},
		"e00": {"", "d02"},
	}

	if !assert.Equal(5, len(nodes), "node count") {
		return
	}
	if !assert.Equal(len(expected), len(names), "event count") {
		return
	}

	index := make(map[EventHash]*Event)
	for _, nodeEvents := range events {
		for _, e := range nodeEvents {
			index[e.Hash()] = e
		}
	}

	for eName, e := range names {
		parents := expected[eName]
		if !assert.Equal(len(parents), len(e.Parents), "at event "+eName) {
			return
		}
		for _, pName := range parents {
			hash := ZeroEventHash
			if pName != "" {
				hash = names[pName].Hash()
			}
			if !e.Parents.Contains(hash) {
				t.Fatalf("%s has no parent %s", eName, pName)
			}
		}
	}
}

// ParseEvents parses events from ASCII-scheme.
// Use joiners ║ ╬ ╠ ╣ ╫ and optional fillers ─ ═ to draw ASCII-scheme.
// Result:
//   - nodes  is an array of node addresses;
//   - events maps node address to array of its events;
//   - names  maps human readable name to the event;
func ParseEvents(asciiScheme string) (
	nodes []common.Address, events map[common.Address][]*Event, names map[string]*Event) {
	// init results
	events = make(map[common.Address][]*Event)
	names = make(map[string]*Event)
	// read lines
	for _, line := range strings.Split(strings.TrimSpace(asciiScheme), "\n") {
		var (
			nNames    []string // event-N --> name
			nCreators []int    // event-N --> creator
			nLinks    [][]int  // event-N --> parents
		)
		// parse line
		current := 0
		for _, symbol := range strings.Split(strings.TrimSpace(line), " ") {
			symbol = strings.TrimSpace(symbol)
			switch symbol {
			case "─", "═", "": // skip filler
				break
			case "╠": // start new link array
				nLinks = append(nLinks, []int{current})
				current++
			case "╬", "╣": // append to last link array
				last := len(nLinks) - 1
				nLinks[last] = append(nLinks[last], current)
				current++
			case "║", "╫": // don't mutate link array
				current++
			default: // it is a event name
				if _, ok := names[symbol]; ok {
					panic(fmt.Errorf("Event '%s' already exists", symbol))
				}
				nCreators = append(nCreators, current)
				nNames = append(nNames, symbol)
				if len(nLinks) < len(nNames) {
					nLinks = append(nLinks, []int(nil))
				}
				current++
			}
		}
		// make nodes if not enough
		for i := len(nodes); i < current; i++ {
			addr := common.FakeAddress()
			nodes = append(nodes, addr)
			events[addr] = nil
		}
		// create events
		for i, name := range nNames {
			// find creator
			creator := nodes[nCreators[i]]
			// find creator's parent
			var (
				parents = EventHashes{}
				ltime   uint64
			)
			if last := len(events[creator]) - 1; last >= 0 {
				parent := events[creator][last]
				parents.Add(parent.Hash())
				ltime = parent.LamportTime
			} else {
				parents.Add(ZeroEventHash)
				ltime = 0
			}
			// find other parents
			for _, p := range nLinks[i] {
				other := nodes[p]
				last := len(events[other]) - 1
				parent := events[other][last]
				parents.Add(parent.Hash())
				if ltime < parent.LamportTime {
					ltime = parent.LamportTime
				}
			}
			// save event
			e := &Event{
				Creator:     creator,
				Parents:     parents,
				LamportTime: ltime + 1,
			}
			events[creator] = append(events[creator], e)
			names[name] = e
			EventNameDict[e.Hash()] = name
		}
	}

	// human readable names for nodes in log
	for node, ee := range events {
		if len(ee) < 1 {
			continue
		}
		name := ee[0].Hash().String()
		common.NodeNameDict[node] = "node" + strings.ToUpper(name[0:1])
	}

	return
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
			parent := ee[len(ee)-1]
			e.Parents.Add(parent.Hash())
			e.LamportTime = parent.LamportTime + 1
		} else {
			e.Parents.Add(ZeroEventHash)
			e.LamportTime = 1
		}
		// other parents are the lasts other's events
		others := parentCount
		for _, other := range parents[1:] {
			if others--; others < 0 {
				break
			}
			if ee := events[nodes[other]]; len(ee) > 0 {
				parent := ee[len(ee)-1]
				e.Parents.Add(parent.Hash())
				if e.LamportTime <= parent.LamportTime {
					e.LamportTime = parent.LamportTime + 1
				}
			}
		}
		// save event
		events[creator] = append(events[creator], e)
	}

	return
}
