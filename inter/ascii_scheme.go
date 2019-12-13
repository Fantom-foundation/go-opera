package inter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

type ForEachEvent struct {
	Process func(e *Event, name string)
	Build   func(e *Event, name string) *Event
}

// ASCIIschemeToDAG parses events from ASCII-scheme for test purpose.
// Use joiners ║ ╬ ╠ ╣ ╫ ╚ ╝ ╩ and optional fillers ─ ═ to draw ASCII-scheme.
// Result:
//   - nodes  is an array of node addresses;
//   - events maps node address to array of its events;
//   - names  maps human readable name to the event;
func ASCIIschemeForEach(
	scheme string,
	callback ForEachEvent,
) (
	nodes []idx.StakerID,
	events map[idx.StakerID][]*Event,
	names map[string]*Event,
) {
	events = make(map[idx.StakerID][]*Event)
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
		prevRef := 0
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
			// do not mark it as new column. Did it on next step.
			if symbol != "╚" && symbol != "╝" {
				col++
			} else {
				// for link with fork
				if ref, ok := prevFarRefs[col]; ok {
					prevRef = ref - 1
				} else {
					prevRef = 1
				}
			}
		}

		for i, name := range nNames {
			// make node if don't exist
			if len(nodes) <= nCreators[i] {
				validator := idx.BytesToStakerID(hash.Of([]byte(name)).Bytes()[:4])
				nodes = append(nodes, validator)
				events[validator] = nil
			}
			// find creator
			creator := nodes[nCreators[i]]
			// find creator's parent
			var (
				index      idx.Event
				parents    = hash.Events{}
				maxLamport idx.Lamport
			)
			if last := len(events[creator]) - prevRef - 1; last >= 0 {
				parent := events[creator][last]
				index = parent.Seq + 1
				parents.Add(parent.Hash())
				maxLamport = parent.Lamport
			} else {
				index = 1
				maxLamport = 0
			}
			// find other parents
			for i, ref := range nLinks[i] {
				if ref < 1 {
					continue
				}
				other := nodes[i]
				last := len(events[other]) - ref
				// fork first event -> Don't add any parents.
				if last < 0 {
					break
				}
				parent := events[other][last]
				if parents.Set().Contains(parent.Hash()) {
					continue
				}
				parents.Add(parent.Hash())
				if maxLamport < parent.Lamport {
					maxLamport = parent.Lamport
				}
			}
			// new event
			e := NewEvent()
			e.Seq = index
			e.Creator = creator
			e.Parents = parents
			e.Lamport = maxLamport + 1
			e.Extra = []byte(name)
			// buildEvent callback
			if callback.Build != nil {
				e = callback.Build(e, name)
			}
			if e == nil {
				continue
			}
			// calc hash of the event, after it's fully built
			e.RecacheHash()
			e.RecacheSize()
			// save event
			events[creator] = append(events[creator], e)
			names[name] = e
			hash.SetEventName(e.Hash(), name)
			// callback
			if callback.Process != nil {
				callback.Process(e, name)
			}
		}
	}

	// human readable names for nodes in log
	for node, ee := range events {
		if len(ee) < 1 {
			continue
		}
		name := []rune(ee[0].Hash().String())
		if strings.HasPrefix(string(name), "node") {
			hash.SetNodeName(node, "node"+strings.ToUpper(string(name[4:5])))
		} else {
			hash.SetNodeName(node, "node"+strings.ToUpper(string(name[0:1])))
		}
	}

	return
}

func ASCIIschemeToDAG(
	scheme string,
) (
	nodes []idx.StakerID,
	events map[idx.StakerID][]*Event,
	names map[string]*Event,
) {
	return ASCIIschemeForEach(scheme, ForEachEvent{})
}

// DAGtoASCIIscheme builds ASCII-scheme of events for debug purpose.
func DAGtoASCIIscheme(events Events) (string, error) {
	events = events.ByParents()

	var (
		scheme rows

		processed = make(map[hash.Event]*Event)
		nodeCols  = make(map[idx.StakerID]int)
		ok        bool

		eventIndex       = make(map[idx.StakerID]map[hash.Event]int)
		creatorLastIndex = make(map[idx.StakerID]int)

		seqCount = make(map[idx.StakerID]map[idx.Event]int)
	)
	for _, e := range events {
		// if count of unique seq > 1 -> fork
		if _, exist := seqCount[e.Creator]; !exist {
			seqCount[e.Creator] = map[idx.Event]int{}
			eventIndex[e.Creator] = map[hash.Event]int{}
		}
		if _, exist := seqCount[e.Creator][e.Seq]; !exist {
			seqCount[e.Creator][e.Seq] = 1
		} else {
			seqCount[e.Creator][e.Seq]++
		}

		if _, exist := creatorLastIndex[e.Creator]; !exist {
			creatorLastIndex[e.Creator] = 0
		} else {
			creatorLastIndex[e.Creator]++
		}

		ehash := e.Hash()
		r := &row{}
		// creator
		if r.Self, ok = nodeCols[e.Creator]; !ok {
			r.Self = len(nodeCols)
			nodeCols[e.Creator] = r.Self
		}
		// name
		r.Name = hash.GetEventName(ehash)
		if len(r.Name) < 1 {
			r.Name = hash.GetNodeName(e.Creator)
			if len(r.Name) < 1 {
				r.Name = string('a' + r.Self)
			}
			r.Name = fmt.Sprintf("%s%03d", r.Name, e.Seq)
		}
		if w := len([]rune(r.Name)); scheme.ColWidth < w {
			scheme.ColWidth = w
		}
		// parents
		r.Refs = make([]int, len(nodeCols))
		selfRefs := 0
		for _, p := range e.Parents {
			parent := processed[p]
			if parent == nil {
				return "", fmt.Errorf("parent %s of %s not found", p.String(), ehash.String())
			}
			if parent.Creator == e.Creator {
				selfRefs++

				// if more then 1 -> fork. Don't skip refs filling.
				if seqCount[e.Creator][e.Seq] == 1 {
					continue
				}
			}

			refCol := nodeCols[parent.Creator]

			var shift int
			if parent.Creator != e.Creator {
				shift = 1
			}

			r.Refs[refCol] = creatorLastIndex[parent.Creator] - eventIndex[parent.Creator][parent.Hash()] + shift
		}
		if (e.Seq <= 1 && selfRefs != 0) || (e.Seq > 1 && selfRefs != 1) {
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

		eventIndex[e.Creator][ehash] = creatorLastIndex[e.Creator]
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

	position byte
)

const (
	none  position = 0
	pass           = iota
	first          = iota
	left           = iota
	right          = iota
	last           = iota
)

func (r *row) Position(i int) position {
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

// Note: after any changes below, run:
// go test -count=100 -run="TestDAGtoASCIIschemeRand" ./inter
// go test -count=100 -run="TestDAGtoASCIIschemeOptimisation" ./inter
func (rr *rows) Optimize() {

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

			// update refs for swapped event (to current event only)
			if len(rr.rows[prev].Refs) > rr.rows[curr].Self {
				// if regression or empty ref
				if rr.rows[prev].Refs[rr.rows[curr].Self] != 1 {
					row.Refs[iRef] = ref
					continue REFS
				}

				rr.rows[prev].Refs[rr.rows[curr].Self]++
			}

			iter := prev + 1
			// update remaining refs for prev event (for events after prev but before curr)
			for pRef, v := range rr.rows[prev].Refs {
				// Note: ref to curr event already updated above.
				if iter == curr {
					break
				}

				// skip self or empty refs
				if pRef == rr.rows[prev].Self || v == 0 {
					continue
				}

				// if next event (after prev but before curr) have refs to prev -> discard swap prev and curr event.
				for nRef := range rr.rows[iter].Refs {
					if nRef == rr.rows[prev].Self {
						row.Refs[iRef] = ref
						continue REFS
					}
				}

				// update remaining refs
				for {
					if pRef == rr.rows[iter].Self && rr.rows[prev].Refs[pRef] < 2 {
						rr.rows[prev].Refs[pRef]++

						// update current prev ref & reset iter for next prev ref
						iter = prev + 1
						break
					}

					if iter < curr {
						iter++
						continue
					}

					// reset iter for next prev ref
					iter = prev + 1
					break
				}
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

			// update index for current event
			curr = prev
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
			if i == row.Self && ref == 0 {
				out(" " + row.Name)
				tail := rr.ColWidth - len([]rune(row.Name)) + 1
				if row.Position(i) == right {
					out(link(tail))
				} else {
					out(nolink(tail))
				}
				continue
			}

			if i == row.Self && ref > 1 {
				tail := rr.ColWidth - len([]rune(row.Name))
				switch row.Position(i) {
				case first:
					out(row.Name + " ╝" + link(tail))
				case last:
					out("╚ " + row.Name + nolink(tail))
				default:
					out("╚ " + row.Name + link(tail))
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
					out("╩╫─" + link(rr.ColWidth-1))
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
