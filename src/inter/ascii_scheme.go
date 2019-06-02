package inter

import (
	"fmt"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// ASCIIschemeToDAG parses events from ASCII-scheme for test purpose.
// Use joiners ║ ╬ ╠ ╣ ╫ ╚ ╝ ╩ and optional fillers ─ ═ to draw ASCII-scheme.
// Result:
//   - nodes  is an array of node addresses;
//   - events maps node address to array of its events;
//   - names  maps human readable name to the event;
func ASCIIschemeToDAG(scheme string) (
	nodes []hash.Peer, events map[hash.Peer][]*Event, names map[string]*Event) {
	// init results
	events = make(map[hash.Peer][]*Event)
	names = make(map[string]*Event)
	// read lines
	for _, line := range strings.Split(strings.TrimSpace(scheme), "\n") {
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
					panic(fmt.Errorf("event '%s' already exists", symbol))
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

type pos byte

const (
	none  pos = 0
	pass      = iota
	first     = iota
	left      = iota
	right     = iota
	last      = iota
)

type schemeRow struct {
	Name  string
	Refs  []int
	Self  int
	First int
	Last  int
}

func (row *schemeRow) Position(i int) pos {
	if i < row.Self {
		// left
		if i < row.First {
			return none
		}
		if i > row.First {
			if row.Refs[i] > 0 {
				return left
			}
			return pass
		}
		return first

	} else {
		// right
		if i > row.Last {
			return none
		}
		if i < row.Last {
			if row.Refs[i] > 0 || i == row.Self {
				return right
			}
			return pass
		}
		return last
	}
}

// DAGtoASCIIcheme builds ASCII-scheme of events for debug purpose.
func DAGtoASCIIcheme(events Events) (string, error) {
	events = events.ByParents()

	// step 1: events to scheme rows
	var (
		scheme []*schemeRow

		processed     = make(map[hash.Event]*Event)
		peerLastIndex = make(map[hash.Peer]uint64)
		peerCols      = make(map[hash.Peer]int)
		colWidth      int
		ok            bool
	)
	for _, e := range events {
		ehash := e.Hash()
		row := &schemeRow{}
		// creator
		if row.Self, ok = peerCols[e.Creator]; !ok {
			row.Self = len(peerCols)
			peerCols[e.Creator] = row.Self
		}
		// name
		row.Name = hash.EventNameDict[ehash]
		if len(row.Name) < 1 {
			row.Name = hash.NodeNameDict[e.Creator]
			if len(row.Name) < 1 {
				row.Name = string('a' + row.Self)
			}
			row.Name = fmt.Sprintf("%s%03d", row.Name, e.Index)
		}
		if colWidth < len(row.Name) {
			colWidth = len(row.Name)
		}
		// parents
		row.Refs = make([]int, len(peerCols))
		selfRefs := 0
		for p := range e.Parents {
			if p.IsZero() {
				selfRefs++
				continue
			}
			parent := processed[p]
			if parent == nil {
				return "", fmt.Errorf("parent %s of %s not found", p.String(), ehash.String())
			}
			if parent.Creator == e.Creator {
				selfRefs++
				continue
			}
			refCol := peerCols[parent.Creator]
			row.Refs[refCol] = int(peerLastIndex[parent.Creator] - parent.Index + 1)
		}
		if selfRefs != 1 {
			return "", fmt.Errorf("self-parents count of %s is %d", ehash, selfRefs)
		}
		// first and last refs
		row.First = len(row.Refs)
		for i, ref := range row.Refs {
			if ref == 0 {
				continue
			}
			if row.First > i {
				row.First = i
			}
			if row.Last < i {
				row.Last = i
			}
		}
		// processed
		scheme = append(scheme, row)
		processed[ehash] = e
		peerLastIndex[e.Creator] = e.Index
	}

	// step 2: scheme rows to strings
	var (
		res strings.Builder
		out = func(s string) {
			_, err := res.WriteString(s)
			if err != nil {
				panic(err)
			}
		}
	)
	colWidth += 3
	for _, row := range scheme {

		// 1st line:
		for i, ref := range row.Refs {
			s := " ║"
			if ref == 2 {
				switch row.Position(i) {
				case first, left:
					s = " ║║"
				case right, last:
					s = "║║"
				}
			}
			if ref > 2 {
				switch row.Position(i) {
				case first, left:
					s = fmt.Sprintf(" ║%d", ref)
				case right, last:
					s = fmt.Sprintf("%d║", ref)
				}
			}
			out(s + nolink(colWidth-len([]rune(s))+2))
		}
		out("\n")

		// 2nd line:
		for i, ref := range row.Refs {
			if i == row.Self {
				out(" " + row.Name)
				tail := colWidth - len(row.Name) + 1
				if row.Position(i) == right {
					out(link(tail))
				} else {
					out(nolink(tail))
				}
				continue
			}

			if ref > 1 {
				switch row.Position(i) {
				case first:
					out(" ║╚" + link(colWidth-1))
				case last:
					out("╝║" + nolink(colWidth))
				case left:
					out(" ╫╩" + link(colWidth-1))
				case right:
					out("╩╫" + link(colWidth))
				case pass:
					out(" ╫" + link(colWidth))
				default:
					out(" ║" + nolink(colWidth))
				}
			} else {
				switch row.Position(i) {
				case first:
					out(" ╠" + link(colWidth))
				case last:
					out(" ╣" + nolink(colWidth))
				case left, right:
					out(" ╬" + link(colWidth))
				case pass:
					out(" ╫" + link(colWidth))
				default:
					out(" ║" + nolink(colWidth))
				}
			}
		}
		out("\n")

	}

	return res.String(), nil
}

func nolink(n int) string {
	return strings.Repeat(" ", n)
}

func link(n int) string {
	if n < 3 {
		return strings.Repeat(" ", n)
	}

	str := strings.Repeat(" ─", (n-1)/2) + " "

	if n%2 == 0 {
		str = str + " "
	}

	return str
}
