package posposet

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

func TestParseEvents(t *testing.T) {
	assert := assert.New(t)

	nodes, events, names := ParseEvents(`
a00 b00 c00 d00
║   ║   ║   ║
a01 ║   ║   ║
║   ╠ ─ c01 ║
a02 ╣   ║   ║
║   ║   ║   ║
╠ ─ ╫ ─ c02 ║
║   b01 ║   ║
║   ╠ ─ ╫ ─ d01
║   ║   ║   ║
║   ║   ║   ║
╠ ─ b02 ╬ ─ ╣
║   ║   ║   ║
a03 ╣   ╠ ─ d02
║   ║   ║   ║
║   ║   ║   ╠ ─ e00
║   ║   ║   ║   ║
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
		if !assert.Equal(len(parents), e.Parents.Len(), "at event "+eName) {
			return
		}
		for _, pName := range parents {
			hash := EventHash{}
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
			var parents EventHashes
			if last := len(events[creator]) - 1; last >= 0 {
				parents.Add(events[creator][last].Hash())
			} else {
				parents.Add(EventHash{})
			}
			// find other parents
			for _, p := range nLinks[i] {
				other := nodes[p]
				last := len(events[other]) - 1
				parents.Add(events[other][last].Hash())
			}
			// save event
			e := &Event{
				Creator: creator,
				Parents: parents,
			}
			events[creator] = append(events[creator], e)
			names[name] = e
		}
	}

	return
}
