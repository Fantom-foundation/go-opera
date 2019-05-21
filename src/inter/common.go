package inter

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"

	"github.com/pkg/errors"

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

type asciiScheme struct {
	graph [][]string

	nodes     map[hash.Peer]uint64
	nodesName map[hash.Peer]rune
	posNodes  []hash.Peer

	eventsPosition map[hash.Event][2]uint64

	lengthColumn uint64
	nextNodeName rune
}

func (scheme *asciiScheme) Len() int {
	return len(scheme.nodes)
}

func (scheme *asciiScheme) Less(i, j int) bool {
	return bytes.Compare(scheme.posNodes[i].Bytes(), scheme.posNodes[j].Bytes()) == -1
}

func (scheme *asciiScheme) Swap(i, j int) {
	for key, pos := range scheme.eventsPosition {
		switch pos[0] {
		case uint64(i):
			pos[0] = uint64(j)
		case uint64(j):
			pos[0] = uint64(i)
		default:
			continue
		}

		scheme.eventsPosition[key] = pos
	}

	scheme.nodes[scheme.posNodes[i]] = uint64(j)
	scheme.nodes[scheme.posNodes[j]] = uint64(i)

	scheme.posNodes[i], scheme.posNodes[j] = scheme.posNodes[j], scheme.posNodes[i]

	scheme.graph[i], scheme.graph[j] = scheme.graph[j], scheme.graph[i]
}

// func (scheme *asciiScheme) calculateLengthColumn()  {
// 	for column := 0; column < len(scheme.graph); column++ {
// 		currentLengthColumn := uint64(len(scheme.graph[column]))
// 		if currentLengthColumn > scheme.lengthColumn {
// 			scheme.lengthColumn = currentLengthColumn
// 		}
// 	}
// }

func (scheme *asciiScheme) insertColumn(after uint64) {
	column := make([]string, scheme.lengthColumn)

	if after >= uint64(len(scheme.graph)) {
		for after >= uint64(len(scheme.graph)) {
			scheme.graph = append(scheme.graph, column)
		}
		return
	}

	after++
	scheme.graph = append(
		scheme.graph[:after],
		append([][]string{column}, scheme.graph[after:]...)...)
}

func (scheme *asciiScheme) insertRow(after uint64) {
	for key := range scheme.eventsPosition {
		pos := scheme.eventsPosition[key]
		if pos[1] <= after {
			continue
		}

		pos[1]++
		scheme.eventsPosition[key] = pos
	}

	after++
	for column := 0; column < len(scheme.graph); column++ {
		var symbol string
		if after >= uint64(len(scheme.graph[column])) {
			scheme.graph[column] = append(scheme.graph[column], symbol)
			continue
		}

		symbol = scheme.graph[column][after-1]
		switch symbol {
		case "║", "╫", "╠", "╣":
			symbol = "║"
		default:
			switch scheme.graph[column][after] {
			case "║", "╫", "╠", "╣":
				symbol = "║"
			default:
				symbol = ""
			}
		}

		scheme.graph[column] = append(
			scheme.graph[column][:after],
			append([]string{symbol}, scheme.graph[column][after:]...)...)
	}
}

func (scheme *asciiScheme) EventsConnect(child, parent hash.Event) {
	if parent == hash.ZeroEvent {
		return
	}

	from := scheme.GetEventPosition(parent)
	to := scheme.GetEventPosition(child)

	if from[0] == to[0] {
		start := from[1]
		stop := to[1]
		column := from[0]

		if from[1] > to[1] {
			start = to[1]
			stop = from[1]
		}

		if stop-start == 1 {
			scheme.insertRow(start)
			scheme.lengthColumn++
			stop++
		}

		start++
		for start < stop {
			scheme.graph[column][start] = "║"
			start++
		}
		return
	}

}

var firstNodeName = rune(97)

func (scheme *asciiScheme) AddEvent(name string, event *Event) {
	column, ok := scheme.nodes[event.Creator]
	if !ok {
		var nextNodeAfter uint64
		if uint64(len(scheme.graph)) != 0 {
			nextNodeAfter = uint64(len(scheme.graph)) - 1
		}
		scheme.insertColumn(nextNodeAfter)
		column = uint64(len(scheme.graph)) - 1
		if scheme.nodes == nil {
			scheme.nodes = make(map[hash.Peer]uint64)
		}
		scheme.nodes[event.Creator] = column
		scheme.posNodes = append(scheme.posNodes, event.Creator)

		if scheme.nextNodeName == 0 {
			scheme.nextNodeName = firstNodeName
		}
		if scheme.nodesName == nil {
			scheme.nodesName = make(map[hash.Peer]rune)
		}
		scheme.nodesName[event.Creator] = scheme.nextNodeName
		scheme.nextNodeName++
	}

	for uint64(len(scheme.graph[column])) <= scheme.lengthColumn {
		scheme.insertRow(scheme.lengthColumn)
	}

	if len(name) == 0 {
		name = fmt.Sprintf("%s%d", string(scheme.nodesName[event.Creator]), event.Index-1)
	}

	scheme.graph[column][scheme.lengthColumn] = name
	if scheme.eventsPosition == nil {
		scheme.eventsPosition = make(map[hash.Event][2]uint64)
	}
	scheme.eventsPosition[event.Hash()] = [2]uint64{column, scheme.lengthColumn}

	scheme.lengthColumn++
}

func (scheme *asciiScheme) GetEventPosition(event hash.Event) [2]uint64 {
	position, ok := scheme.eventsPosition[event]
	if !ok {
		panic(errors.New("can't find event"))
	}
	return position
}

func (scheme *asciiScheme) String() string {
	return ""
}

func CreateSchemaByEvents(events Events) string {
	events = events.ByParents()

	scheme := new(asciiScheme)

	for _, event := range events {
		scheme.AddEvent("", event)
		for parent := range event.Parents {
			if parent == hash.ZeroEvent {
				continue
			}
			scheme.EventsConnect(event.Hash(), parent)
		}
	}

	return scheme.String()
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
