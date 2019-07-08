package inter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// ASCIIschemeToDAG parses events from ASCII-scheme for test purpose.
// Use joiners ║ ╬ ╠ ╣ ╫ ╚ ╝ ╩ and optional fillers ─ ═ to draw ASCII-scheme.
// Result:
//   - nodes  is an array of node addresses;
//   - events maps node address to array of its events;
//   - names  maps human readable name to the event;
func ASCIIschemeToDAG(scheme string) (
	nodes []hash.Peer,
	events map[hash.Peer][]*Event,
	names map[string]*Event,
) {
	events = make(map[hash.Peer][]*Event)
	names = make(map[string]*Event)
	var (
		prevFarRefs map[int]int
		curFarRefs  = make(map[int]int)
	)
	// read lines
	for _, line := range strings.Split(strings.TrimSpace(scheme), "\n") {
		var (
			nNames    []string // event-N --> name
			nCreators []int    // event-N --> creator
			nLinks    [][]int  // event-N --> n-parent ref (0 if not)
		)
		prevFarRefs, curFarRefs = curFarRefs, make(map[int]int)

		// parse line
		col := 0
		for _, symbol := range strings.FieldsFunc(strings.TrimSpace(line), filler) {
			symbol = strings.TrimSpace(symbol)
			if strings.HasPrefix(symbol, "//") {
				break // skip comments
			}

			switch symbol {
			case "": // skip
				col--
			case "╠", "║╠", "╠╫": // start new link array with current
				refs := make([]int, col+1)
				refs[col] = 1
				nLinks = append(nLinks, refs)
			case "║╚", "╚": // start new link array with prev
				refs := make([]int, col+1)
				if ref, ok := prevFarRefs[col]; ok {
					refs[col] = ref
				} else {
					refs[col] = 2
				}
				nLinks = append(nLinks, refs)
			case "╣", "╣║", "╫╣", "╬": // append current to last link array
				last := len(nLinks) - 1
				nLinks[last] = append(nLinks[last], make([]int, col+1-len(nLinks[last]))...)
				nLinks[last][col] = 1
			case "╝║", "╝", "╩╫", "╫╩": // append prev to last link array
				last := len(nLinks) - 1
				nLinks[last] = append(nLinks[last], make([]int, col+1-len(nLinks[last]))...)
				if ref, ok := prevFarRefs[col]; ok {
					nLinks[last][col] = ref
				} else {
					nLinks[last][col] = 2
				}
			case "╫", "║", "║║": // don't mutate link array
				break
			default:
				if strings.HasPrefix(symbol, "║") || strings.HasSuffix(symbol, "║") {
					// it is a far ref
					symbol = strings.Trim(symbol, "║")
					ref, err := strconv.ParseInt(symbol, 10, 64)
					if err != nil {
						panic(err)
					}
					curFarRefs[col] = int(ref)
				} else {
					// it is a event name
					if _, ok := names[symbol]; ok {
						panic(fmt.Errorf("event '%s' already exists", symbol))
					}
					nCreators = append(nCreators, col)
					nNames = append(nNames, symbol)
					if len(nLinks) < len(nNames) {
						refs := make([]int, col+1)
						nLinks = append(nLinks, refs)
					}
				}
			}
			col++
		}
		// make nodes if not enough
		for i := len(nodes); i < (col); i++ {
			addr := hash.FakePeer(int64(i + 1))
			nodes = append(nodes, addr)
			events[addr] = nil
		}
		// create events
		for i, name := range nNames {
			// find creator
			creator := nodes[nCreators[i]]
			// find creator's parent
			var (
				index      idx.Event
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
			for i, ref := range nLinks[i] {
				if ref < 1 {
					continue
				}
				other := nodes[i]
				last := len(events[other]) - ref
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

// DAGtoASCIIscheme builds ASCII-scheme of events for debug purpose.
func DAGtoASCIIscheme(events Events) (string, error) {
	events = events.ByParents()

	var (
		scheme rows

		processed     = make(map[hash.Event]*Event)
		peerLastIndex = make(map[hash.Peer]idx.Event)
		peerCols      = make(map[hash.Peer]int)
		ok            bool
	)
	for _, e := range events {
		ehash := e.Hash()
		r := &row{}
		// creator
		if r.Self, ok = peerCols[e.Creator]; !ok {
			r.Self = len(peerCols)
			peerCols[e.Creator] = r.Self
		}
		// name
		r.Name = hash.EventNameDict[ehash]
		if len(r.Name) < 1 {
			r.Name = hash.NodeNameDict[e.Creator]
			if len(r.Name) < 1 {
				r.Name = string('a' + r.Self)
			}
			r.Name = fmt.Sprintf("%s%03d", r.Name, e.Index)
		}
		if w := len([]rune(r.Name)); scheme.ColWidth < w {
			scheme.ColWidth = w
		}
		// parents
		r.Refs = make([]int, len(peerCols))
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
			r.Refs[refCol] = int(peerLastIndex[parent.Creator] - parent.Index + 1)
		}
		if selfRefs != 1 {
			return "", fmt.Errorf("self-parents count of %s is %d", ehash, selfRefs)
		}
		// first and last refs
		r.First = len(r.Refs)
		for i, ref := range r.Refs {
			if ref == 0 {
				continue
			}
			if r.First > i {
				r.First = i
			}
			if r.Last < i {
				r.Last = i
			}
		}
		// processed
		scheme.Add(r)
		processed[ehash] = e
		peerLastIndex[e.Creator] = e.Index
	}

	scheme.Optimize()

	scheme.ColWidth += 3
	return scheme.String(), nil
}

func filler(r rune) bool {
	return r == ' ' || r == '─' || r == '═'
}

/*
 * staff:
 */

type (
	row struct {
		Name  string
		Refs  []int
		Self  int
		First int
		Last  int
	}

	rows struct {
		rows     []*row
		ColWidth int
	}

	pos byte
)

const (
	none  pos = 0
	pass      = iota
	first     = iota
	left      = iota
	right     = iota
	last      = iota
)

func (r *row) Position(i int) pos {
	// if left
	if i < r.Self {
		if i < r.First {
			return none
		}
		if i > r.First {
			if r.Refs[i] > 0 {
				return left
			}
			return pass
		}
		return first
	}
	// else right
	if i > r.Last {
		return none
	}
	if i < r.Last {
		if r.Refs[i] > 0 || i == r.Self {
			return right
		}
		return pass
	}
	return last
}

func (rr *rows) Optimize() {
	// TODO: fix than uncomment
	// NOTE: see `go test -count=100 -run="TestDAGtoASCIIschemeRand" ./src/inter`
	return

	for curr, row := range rr.rows {
	REFS:
		for iRef, ref := range row.Refs {
			// TODO: Can we decrease ref from 2 to 1 ?
			if ref < 3 {
				continue REFS
			}

			// find prev event for swap
			prev := curr - 1
			for {
				if rr.rows[prev].Self == iRef {
					break
				}
				// if the same parents
				if rr.rows[curr].Self == rr.rows[prev].Self {
					continue REFS
				}

				prev--
			}

			row.Refs[iRef] = ref - 1

			// update refs for swapped event
			if rr.rows[prev].Refs[rr.rows[curr].Self] > 0 {
				rr.rows[prev].Refs[rr.rows[curr].Self]++
			}

			// Note: if swapped event now have refs more than before -> discard any changes for current and swapped event.
			if rr.rows[prev].Refs[rr.rows[curr].Self] > 2 {
				rr.rows[prev].Refs[rr.rows[curr].Self]--
				row.Refs[iRef] = ref
				continue
			}

			// for fill empty space after swap (for graph)
			for {
				if len(rr.rows[prev].Refs) == len(rr.rows[curr].Refs) {
					break
				}

				rr.rows[prev].Refs = append(rr.rows[prev].Refs, 0)
			}

			// swap with prev event
			rr.rows[curr], rr.rows[prev] = rr.rows[prev], rr.rows[curr]
		}
	}
}

func (rr *rows) String() string {
	var (
		res strings.Builder
		out = func(s string) {
			_, err := res.WriteString(s)
			if err != nil {
				panic(err)
			}
		}
	)
	for _, row := range rr.rows {

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
			if ref > 2 { // far ref
				switch row.Position(i) {
				case first, left:
					s = fmt.Sprintf(" ║%d", ref)
				case right, last:
					s = fmt.Sprintf("%d║", ref)
				}
			}
			out(s + nolink(rr.ColWidth-len([]rune(s))+2))
		}
		out("\n")

		// 2nd line:
		for i, ref := range row.Refs {
			if i == row.Self {
				out(" " + row.Name)
				tail := rr.ColWidth - len([]rune(row.Name)) + 1
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
					out(" ║╚" + link(rr.ColWidth-1))
				case last:
					out("╝║" + nolink(rr.ColWidth))
				case left:
					out("─╫╩" + link(rr.ColWidth-1))
				case right:
					out("╩╫─" + link(rr.ColWidth))
				case pass:
					out("─╫─" + link(rr.ColWidth-1))
				default:
					out(" ║" + nolink(rr.ColWidth))
				}
			} else {
				switch row.Position(i) {
				case first:
					out(" ╠" + link(rr.ColWidth))
				case last:
					out("═╣" + nolink(rr.ColWidth))
				case left, right:
					out("═╬" + link(rr.ColWidth))
				case pass:
					out("─╫─" + link(rr.ColWidth-1))
				default:
					out(" ║" + nolink(rr.ColWidth))
				}
			}
		}
		out("\n")

	}
	return res.String()
}

func (rr *rows) Add(r *row) {
	rr.rows = append(rr.rows, r)
}

func nolink(n int) string {
	return strings.Repeat(" ", n)
}

func link(n int) string {
	if n < 3 {
		return strings.Repeat(" ", n)
	}

	str := strings.Repeat("══", (n-1)/2) + "═"

	if n%2 == 0 {
		str = str + "═"
	}

	return str
}
