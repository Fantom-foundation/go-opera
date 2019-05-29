package inter

import (
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

// asciiScheme helping type for create ascii scheme by events
type asciiScheme struct {
	graph [][]string

	nodes     map[hash.Peer]uint64
	nodesName map[hash.Peer]rune

	eventsPosition map[hash.Event][2]uint64

	lengthColumn uint64
	nextNodeName rune
}

// increaseEventsPositions add offset for events positions after insert row or column.
// For insert row 'index' must be 0.
// For insert column 'index' must be 1.
func (scheme *asciiScheme) increaseEventsPositions(after uint64, index int) {
	for key := range scheme.eventsPosition {
		pos := scheme.eventsPosition[key]
		if pos[index] <= after {
			continue
		}

		pos[index]++
		scheme.eventsPosition[key] = pos
	}
}

// insertColumn insert column after specific column ('after' parameter).
func (scheme *asciiScheme) insertColumn(after uint64) {
	scheme.increaseEventsPositions(after, 0)

	for node, column := range scheme.nodes {
		if column > after {
			scheme.nodes[node] = column + 1
		}
	}

	column := make([]string, scheme.lengthColumn)

	if after >= uint64(len(scheme.graph)) {
		lastColumn := len(scheme.graph) - 1
		if lastColumn >= 0 {
			for i := 0; i < len(column); i++ {
				switch scheme.graph[lastColumn][i] {
				case "╠", "╫", "╚", "╩":
					column[i] = "-"
				}
			}
		}

		for after >= uint64(len(scheme.graph)) {
			scheme.graph = append(scheme.graph, column)
		}
		return
	}

	for i := 0; i < len(column); i++ {
		switch scheme.graph[after][i] {
		case "╠", "╫", "╚", "╩", "-":
			column[i] = "-"
			continue
		}

		if after+1 != uint64(len(scheme.graph)) {
			switch scheme.graph[after+1][i] {
			case "╣", "╫", "╝", "╩", "-":
				column[i] = "-"
			}
		}
	}

	after++
	scheme.graph = append(
		scheme.graph[:after],
		append([][]string{column}, scheme.graph[after:]...)...)
}

// insertRow insert row after specific row ('after' parameter).
func (scheme *asciiScheme) insertRow(after uint64) {
	if len(scheme.graph) == 0 {
		return
	}

	scheme.increaseEventsPositions(after, 1)
	scheme.lengthColumn++

	connections := "║╫╠╣╝╚"
	after++
	for column := 0; column < len(scheme.graph); column++ {
		var symbol string
		if after >= uint64(len(scheme.graph[column])) {
			scheme.graph[column] = append(scheme.graph[column], symbol)
			continue
		}

		symbol = "║"

		lastSymbol := scheme.graph[column][after-1]
		indexLastSymbol := -1
		if len(lastSymbol) > 0 {
			indexLastSymbol = strings.Index(connections, lastSymbol)
		}

		nextSymbol := scheme.graph[column][after]
		indexNextSymbol := -1
		if len(nextSymbol) > 0 {
			indexNextSymbol = strings.Index(connections, nextSymbol)
		}

		if (indexLastSymbol == -1 && indexNextSymbol != -1) ||
			(indexLastSymbol != -1 && indexNextSymbol == -1) ||
			(indexLastSymbol == -1 && indexNextSymbol == -1) {
			symbol = ""
		}

		scheme.graph[column] = append(
			scheme.graph[column][:after],
			append([]string{symbol}, scheme.graph[column][after:]...)...)
	}
}

func (scheme *asciiScheme) drawVerticalConnector(from, to [2]uint64) {
	start := from[1]
	stop := to[1]
	column := from[0]

	if from[1] > to[1] {
		start = to[1]
		stop = from[1]
	}

	for start <= stop {
		var connector string
		switch scheme.graph[column][start] {
		case "-":
			connector = "╫"
		case "":
			connector = "║"
		case "╝":
			connector = "╣"
		case "╚":
			connector = "╠"
		default:
			start++
			continue
		}

		var leftSymbol, rightSymbol string
		if int64(column-1) >= 0 {
			leftSymbol = scheme.graph[column-1][start]
		}
		if column+1 < uint64(len(scheme.graph)) {
			rightSymbol = scheme.graph[column+1][start]
		}

		if (leftSymbol == "-" || leftSymbol == "╠" || leftSymbol == "╫") &&
			(rightSymbol == "-" || rightSymbol == "╣" || rightSymbol == "╫") {
			connector = "╫"
		} else if leftSymbol == "-" && rightSymbol != "-" {
			connector = "╣"
		} else if leftSymbol != "-" && rightSymbol == "-" {
			connector = "╠"
		}

		scheme.graph[column][start] = connector
		start++
	}
}

func (scheme *asciiScheme) drawHorizontalConnector(from, to [2]uint64) {
	start := from[0]
	stop := to[0]

	isLeft := true

	if from[0] > to[0] {
		start = to[0]
		stop = from[0]
		isLeft = false
	}

	scheme.setVerticalConnector(to[0], from[1], isLeft)
	scheme.setVerticalConnector(from[0], from[1], !isLeft)

	if stop-start == 1 {
		scheme.insertColumn(start)
		stop++
	}

	start++
	for start != stop {
		connector := "-"
		if scheme.graph[start][from[1]] == "║" {
			connector = "╫"
		}
		scheme.graph[start][from[1]] = connector
		start++
	}
}

func (scheme *asciiScheme) setVerticalConnector(column, row uint64, isLeft bool) {
	nodeConnector := "╝"
	if !isLeft {
		// nodeConnector = "╠"
		nodeConnector = "╚"
	}

	switch scheme.graph[column][row] {
	case "╬":
		nodeConnector = "╬"
	case "╣":
		if nodeConnector == "╠" {
			nodeConnector = "╬"
		}
	case "╠":
		if nodeConnector == "╣" {
			nodeConnector = "╬"
		}
	}
	scheme.graph[column][row] = nodeConnector
}

// EventsConnect add communication from child to parent.
// If parent is zero event, do nothing.
func (scheme *asciiScheme) EventsConnect(child, parent hash.Event) {
	if parent == hash.ZeroEvent {
		return
	}

	from := scheme.getEventPosition(child)
	to := scheme.getEventPosition(parent)

	if from[0] == to[0] {
		if int64(from[1]-to[1]) == 1 {
			scheme.insertRow(to[1])
			to[1]++
		} else if int64(to[1]-from[1]) == 1 {
			scheme.insertRow(from[1])
			from[1]++
		}

		scheme.drawVerticalConnector(from, to)
		return
	}

	scheme.insertRow(from[1])
	from[1]++
	connectorRow := from[1]

	var isDrawFromVerticalCommunication bool

	for event, pos := range scheme.eventsPosition {
		if event == child {
			continue
		}

		if to[0] == pos[0] && to[1] < pos[1] {
			scheme.insertColumn(to[0])
			from[0]++
			to[0]++
			scheme.drawVerticalConnector(to, [2]uint64{to[0], from[1]})
			isDrawFromVerticalCommunication = true

			if scheme.graph[to[0]][to[1]] == "╫" {
				scheme.graph[to[0]][to[1]] = "-"
				break
			}

			scheme.graph[to[0]][to[1]] = ""
			break
		}
	}

	scheme.drawHorizontalConnector(from, to)

	if !isDrawFromVerticalCommunication {
		to = scheme.getEventPosition(parent)
		scheme.drawVerticalConnector([2]uint64{to[0], connectorRow}, [2]uint64{to[0], to[1] + 1})
	}

	from = scheme.getEventPosition(child)
	scheme.drawVerticalConnector([2]uint64{from[0], connectorRow}, [2]uint64{from[0], connectorRow})
}

// initial node name for generating ascii scheme by events (current value 'a')
const firstNodeName = rune(97)

// AddEvent register new event
// If node of event isn't exist, node will create.
// Parameter 'name' is optional (possible be empty). If 'name' is empty, name will generate beginning name node 'a' and
// event index in node.
func (scheme *asciiScheme) AddEvent(name string, event *Event) {
	if event == nil {
		panic(errors.Errorf("event '%s' must be set", name))
	}

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

		if scheme.nextNodeName == 0 {
			scheme.nextNodeName = firstNodeName
		}
		if scheme.nodesName == nil {
			scheme.nodesName = make(map[hash.Peer]rune)
		}
		scheme.nodesName[event.Creator] = scheme.nextNodeName
		scheme.nextNodeName++
	}

	currentLengthColumn := scheme.lengthColumn

	for uint64(len(scheme.graph[column])) <= currentLengthColumn {
		scheme.insertRow(scheme.lengthColumn)
	}

	if len(name) == 0 {
		name = fmt.Sprintf("%s%d", string(scheme.nodesName[event.Creator]), event.Index-1)
	}

	scheme.graph[column][currentLengthColumn] = name
	if scheme.eventsPosition == nil {
		scheme.eventsPosition = make(map[hash.Event][2]uint64)
	}
	scheme.eventsPosition[event.Hash()] = [2]uint64{column, currentLengthColumn}
}

// getEventPosition return position event in ascii scheme
func (scheme *asciiScheme) getEventPosition(event hash.Event) [2]uint64 {
	position, ok := scheme.eventsPosition[event]
	if !ok {
		panic(errors.New("can't find event"))
	}
	return position
}

// String return ascii scheme on means single string
func (scheme *asciiScheme) String() string {
	var asciiScheme string

	for column := 0; column < len(scheme.graph); column++ {
		var currentLength int
		for row := 0; row < len(scheme.graph[column]); row++ {
			switch scheme.graph[column][row] {
			case "╠", "╣", "╫", "║", "╝", "╩", "╚":
				continue
			default:
				if currentLength < len(scheme.graph[column][row]) {
					currentLength = len(scheme.graph[column][row])
				}
			}
		}

		if currentLength <= 1 {
			continue
		}

		for i := 0; i < currentLength; i++ {
			scheme.insertColumn(uint64(column))
		}
		for _, pos := range scheme.eventsPosition {
			if pos[0] == uint64(column) {
				for i := 1; i < currentLength; i++ {
					scheme.graph[column+i][pos[1]] = "&"
				}

			}
		}
	}

	for row := 0; row < int(scheme.lengthColumn); row++ {
		for column := 0; column < len(scheme.graph); column++ {
			if scheme.graph[column][row] == "&" {
				continue
			}
			symbol := scheme.graph[column][row]
			if len(symbol) == 0 {
				symbol = " "
			}

			asciiScheme += symbol
		}
		asciiScheme += "\n"
	}

	return asciiScheme
}

// CreateSchemaByEvents return ascii scheme by events with parents
// Throw panic:
// - event is nil;
// - not correctly works type asciiScheme.
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
