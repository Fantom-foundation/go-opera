package seeing

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
)

// Strongly is a datas to detect strongly-see condition.
type Strongly struct {
	members internal.Members
	nodes   map[hash.Peer]int
	events  map[hash.Event]*Event

	logger.Instance
}

// New creates Strongly instance.
func New(mm internal.Members) *Strongly {
	ss := &Strongly{
		Instance: logger.MakeInstance(),
	}
	ss.Reset(mm)

	return ss
}

// Reset resets buffers.
func (ss *Strongly) Reset(mm internal.Members) {
	ss.members = mm
	ss.nodes = make(map[hash.Peer]int)
	ss.events = make(map[hash.Event]*Event)
}

func (ss *Strongly) Add(e *inter.Event) {
	// sanity check
	if _, ok := ss.events[e.Hash()]; ok {
		ss.Fatalf("event %s already exists", e.Hash().String())
		return
	}

	event := &Event{
		Event:       e,
		LowestSees:  make([]idx.Event, len(ss.members)),
		HighestSeen: make([]idx.Event, len(ss.members)),
	}

	ss.setNodes(event)
	ss.fillEventRefs(event)
	ss.events[e.Hash()] = event
}

func (ss *Strongly) setNodes(e *Event) {
	var ok bool
	if e.MemberN, ok = ss.nodes[e.Creator]; !ok {
		e.MemberN = len(ss.nodes)
		ss.nodes[e.Creator] = e.MemberN
	}
}

func (ss *Strongly) fillEventRefs(e *Event) {
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
		if e.HighestSeen[i] < n {
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
	if e.LowestSees[node] == 0 ||
		e.LowestSees[node] > ref {
		e.LowestSees[node] = ref
		return true
	}
	return false
}

// See() calculates "sufficient coherence" between the events.
// The A.HighestSeen array remembers the sequence number of the last
// event by each member that is an ancestor of A. The array for
// B.LowestSees remembers the sequence number of the earliest
// event by each member that is a descendant of B. Compare the two arrays,
// and find how many elements in the A.HighestSeen array are greater
// than or equal to the corresponding element of the B.LowestSees
// array. If there are more than 2n/3 such matches, then the A and B
// have achieved sufficient coherency.
func (ss *Strongly) See(aHash, bHash hash.Event) bool {
	counter := ss.members.NewCounter()

	// get events by hash
	a, ok := ss.events[aHash]
	if !ok {
		return false
	}
	b, ok := ss.events[bHash]
	if !ok {
		return false
	}

	// calculate strongly seeing using the indexes
	for m := range ss.members {
		n := ss.nodes[m]
		if b.LowestSees[n] <= a.HighestSeen[n] && b.LowestSees[n] != 0 {
			counter.Count(m)
		}
	}

	return counter.HasQuorum()
}
