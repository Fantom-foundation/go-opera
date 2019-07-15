package seeing

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
)

const (
	nodeCount = 512
)

// Strongly is a datas to detect strongly-see condition.
type Strongly struct {
	newCounter internal.StakeCounterProvider
	nodes      map[hash.Peer]int
	events     map[hash.Event]*Event

	logger.Instance
}

// New creates Strongly instance.
func New(c internal.StakeCounterProvider) *Strongly {
	ss := &Strongly{
		newCounter: c,
		Instance:   logger.MakeInstance(),
	}
	ss.Reset()

	return ss
}

// Reset resets buffers.
func (ss *Strongly) Reset() {
	ss.nodes = make(map[hash.Peer]int, nodeCount)
	ss.events = make(map[hash.Event]*Event)
}

// Cache event for Strongly.See.
func (ss *Strongly) Cache(e *inter.Event) {
	// sanity check
	if _, ok := ss.events[e.Hash()]; ok {
		ss.Fatalf("event %s already exists", e.Hash().String())
		return
	}

	event := &Event{
		Event:   e,
		MemberN: ss.nodeIndex(e.Creator),
	}

	ss.fillEventRefs(event)
	ss.events[e.Hash()] = event
}

func (ss *Strongly) nodeIndex(n hash.Peer) int {
	var (
		index int
		ok    bool
	)
	if index, ok = ss.nodes[n]; !ok {
		index = len(ss.nodes)
		ss.nodes[n] = index
	}

	return index
}

func (ss *Strongly) fillEventRefs(e *Event) {
	e.LowestSees = make([]idx.Event, e.MemberN+1, nodeCount)
	e.HighestSeen = make([]idx.Event, e.MemberN+1, nodeCount)

	// seen by himself
	e.LowestSees[e.MemberN] = e.Index
	e.HighestSeen[e.MemberN] = e.Index

	for p := range e.Parents {
		if p.IsZero() {
			continue
		}
		parent := ss.events[p]
		ss.updateAllLowestSees(parent, e.MemberN, e.Index)
		ss.updateAllHighestSeen(e, parent)
	}
}

func (ss *Strongly) updateAllHighestSeen(e, parent *Event) {
	for i, n := range parent.HighestSeen {
		if getRef(&e.HighestSeen, i) < n {
			e.HighestSeen[i] = n
		}
	}
}

func (ss *Strongly) updateAllLowestSees(e *Event, node int, ref idx.Event) {
	toUpdate := []*Event{e}
	for {
		var next []*Event
		for _, event := range toUpdate {
			if !setLowestSeesIfMin(event, node, ref) {
				continue
			}
			for p := range event.Parents {
				if !p.IsZero() {
					next = append(next, ss.events[p])
				}
			}
		}

		if len(next) == 0 {
			break
		}
		toUpdate = next
	}
}

func setLowestSeesIfMin(e *Event, node int, ref idx.Event) bool {
	curr := getRef(&e.LowestSees, node)
	if curr == 0 || curr > ref {
		e.LowestSees[node] = ref
		return true
	}
	return false
}

// See calculates "sufficient coherence" between the events.
// The A.HighestSeen array remembers the sequence number of the last
// event by each member that is an ancestor of A. The array for
// B.LowestSees remembers the sequence number of the earliest
// event by each member that is a descendant of B. Compare the two arrays,
// and find how many elements in the A.HighestSeen array are greater
// than or equal to the corresponding element of the B.LowestSees
// array. If there are more than 2n/3 such matches, then the A and B
// have achieved sufficient coherency.
func (ss *Strongly) See(aHash, bHash hash.Event) bool {
	// get events by hash
	a, ok := ss.events[aHash]
	if !ok {
		return false
	}
	b, ok := ss.events[bHash]
	if !ok {
		return false
	}

	yes := ss.newCounter()
	no := ss.newCounter()

	// calculate strongly seeing using the indexes
	for m, n := range ss.nodes {
		bLowestSees := getRef(&b.LowestSees, n)
		aHighestSeen := getRef(&a.HighestSeen, n)
		if bLowestSees <= aHighestSeen && bLowestSees != 0 {
			yes.Count(m)
		} else {
			no.Count(m)
		}

		if yes.HasQuorum() {
			return true
		}

		if no.HasQuorum() {
			return false
		}
	}

	return false
}

func getRef(rr *[]idx.Event, i int) idx.Event {
	n := len(*rr)
	if n <= i {
		*rr = append(*rr, make([]idx.Event, i-n+1)...)
	}
	return (*rr)[i]
}
