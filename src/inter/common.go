package inter

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// ParseEvents parses events from ASCII-scheme for test purpose.
// Use joiners ║ ╬ ╠ ╣ ╫ ╚ ╝ ╩ and optional fillers ─ ═ to draw ASCII-scheme.
// Result:
//   - nodes  is an array of node addresses;
//   - events maps node address to array of its events;
//   - names  maps human readable name to the event;
func ParseEvents(asciiScheme string) (
	nodes []hash.Peer, events map[hash.Peer][]*Event, names map[string]*Event) {
	// init results
	events = make(map[hash.Peer][]*Event)
	names = make(map[string]*Event)
	// read lines
	for _, line := range strings.Split(strings.TrimSpace(asciiScheme), "\n") {
		var (
			nNames    []string // event-N --> name
			nCreators []int    // event-N --> creator
			nLinks    [][]int  // event-N --> parents+1 (negative if link to pre-last event)
		)
		// parse line
		current := 1
		for _, symbol := range strings.Split(strings.TrimSpace(line), " ") {
			symbol = strings.TrimSpace(symbol)
			switch symbol {
			case "─", "═", "": // skip filler
				current--
			case "╠", "║╠", "╠╫": // start new link array with current
				nLinks = append(nLinks, []int{current})
			case "║╚", "╚": // start new link array with prev
				nLinks = append(nLinks, []int{-1 * current})
			case "╣", "╣║", "╫╣", "╬": // append current to last link array
				last := len(nLinks) - 1
				nLinks[last] = append(nLinks[last], current)
			case "╝║", "╝", "╩╫", "╫╩": // append prev to last link array
				last := len(nLinks) - 1
				nLinks[last] = append(nLinks[last], -1*current)
			case "╫", "║", "║║": // don't mutate link array
				break
			default: // it is a event name
				if _, ok := names[symbol]; ok {
					panic(fmt.Errorf("Event '%s' already exists", symbol))
				}
				nCreators = append(nCreators, current-1)
				nNames = append(nNames, symbol)
				if len(nLinks) < len(nNames) {
					nLinks = append(nLinks, []int(nil))
				}
			}
			current++
		}
		// make nodes if not enough
		for i := len(nodes); i < (current - 1); i++ {
			addr := hash.FakePeer()
			nodes = append(nodes, addr)
			events[addr] = nil
		}
		// create events
		for i, name := range nNames {
			// find creator
			creator := nodes[nCreators[i]]
			// find creator's parent
			var (
				index      uint64
				parents    = hash.Events{}
				maxLamport Timestamp
			)
			if last := len(events[creator]) - 1; last >= 0 {
				parent := events[creator][last]
				index = parent.Index + 1
				parents.Add(parent.Hash())
				maxLamport = parent.LamportTime
			} else {
				index = 1
				parents.Add(hash.ZeroEvent)
				maxLamport = 0
			}
			// find other parents
			for _, p := range nLinks[i] {
				prev := 0
				if p < 0 {
					p *= -1
					prev = -1
				}
				p = p - 1
				other := nodes[p]
				last := len(events[other]) - 1 + prev
				parent := events[other][last]
				parents.Add(parent.Hash())
				if maxLamport < parent.LamportTime {
					maxLamport = parent.LamportTime
				}
			}
			// save event
			e := &Event{
				Index:       index,
				Creator:     creator,
				Parents:     parents,
				LamportTime: maxLamport + 1,
			}
			events[creator] = append(events[creator], e)
			names[name] = e
			hash.EventNameDict[e.Hash()] = name
		}
	}

	// human readable names for nodes in log
	for node, ee := range events {
		if len(ee) < 1 {
			continue
		}
		name := ee[0].Hash().String()
		hash.NodeNameDict[node] = "node" + strings.ToUpper(name[0:1])
	}

	return
}

type schemaEvents [][]string

func (schema schemaEvents) insertColumn(after Timestamp) {
}

func (schema schemaEvents) insertRow(column, after Timestamp) {
	after++
	schema[column] = append(
		schema[column][:after],
		append([]string{""}, schema[column][after:]...)...)
}

func (schema schemaEvents) connect(from, to [2]Timestamp) {
	if from[0] == to[0] {
		start := from[1]
		stop := to[1]
		column := from[0]

		if from[1] > to[1] {
			start = to[1]
			stop = from[1]
		}

		if stop-start == 1 {
			schema.insertRow(from[0], start)
			stop++
		}

		start++
		for start < stop {
			schema[column][start] = "║"
			start++
		}
		return
	}

	println(123)
}

func CreateSchemaByEvents(events map[string]*Event) (asciiSchema string) {
	eventsPos := make(map[hash.Event][2]Timestamp)
	schema := make(schemaEvents, 0)
	nodes := make(map[hash.Peer]Timestamp)

	for key, event := range events {
		column, ok := nodes[event.Creator]
		if !ok {
			schema = append(schema, []string{})
			column = Timestamp(len(schema) - 1)
			nodes[event.Creator] = column
		}

		row := event.LamportTime - 1
		for Timestamp(len(schema[column])) <= row {
			schema[column] = append(schema[column], "")
		}

		schema[column][row] = key
		eventsPos[event.Hash()] = [2]Timestamp{column, row}
	}

	for key, event := range events {
		println(key)

		pos := eventsPos[event.Hash()]

		for e := range event.Parents {
			if e == hash.ZeroEvent {
				continue
			}

			parentPos := eventsPos[e]
			schema.connect(pos, parentPos)
		}
	}

	return
}

// GenEventsByNode generates random events for test purpose.
// Result:
//   - nodes  is an array of node addresses;
//   - events maps node address to array of its events;
func GenEventsByNode(nodeCount, eventCount, parentCount int) (
	nodes []hash.Peer, events map[hash.Peer][]*Event) {
	// init results
	nodes = make([]hash.Peer, nodeCount)
	events = make(map[hash.Peer][]*Event, nodeCount)
	// make and name nodes
	for i := 0; i < nodeCount; i++ {
		addr := hash.FakePeer()
		nodes[i] = addr
		hash.NodeNameDict[addr] = "node" + string('A'+i)
	}
	// make events
	for i := 0; i < nodeCount*eventCount; i++ {
		// seq parent
		self := i % nodeCount
		parents := rand.Perm(nodeCount)
		creator := nodes[self]
		for j, n := range parents {
			if n == self {
				parents = append(parents[0:j], parents[j+1:]...)
				break
			}
		}
		// make
		e := &Event{
			Creator: creator,
			Parents: hash.Events{},
		}
		// first parent is a last creator's event or empty hash
		if ee := events[creator]; len(ee) > 0 {
			parent := ee[len(ee)-1]
			e.Parents.Add(parent.Hash())
			e.LamportTime = parent.LamportTime + 1
		} else {
			e.Parents.Add(hash.ZeroEvent)
			e.LamportTime = 1
		}
		// other parents are the lasts other's events
		for _, other := range parents[1:parentCount] {
			if ee := events[nodes[other]]; len(ee) > 0 {
				parent := ee[len(ee)-1]
				e.Parents.Add(parent.Hash())
				if e.LamportTime <= parent.LamportTime {
					e.LamportTime = parent.LamportTime + 1
				}
			}
		}
		// save and name event
		hash.EventNameDict[e.Hash()] = fmt.Sprintf("%s%03d", string('a'+self), len(events[creator]))
		events[creator] = append(events[creator], e)
	}

	return
}
