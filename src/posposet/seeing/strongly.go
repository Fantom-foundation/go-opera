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
	roots   []map[idx.Frame][]*Event

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

	ss.roots = make([]map[idx.Frame][]*Event, len(mm))
	for i := range ss.roots {
		ss.roots[i] = make(map[idx.Frame][]*Event)
	}
}

func (ss *Strongly) See(a hash.Event, b hash.Peer, f idx.Frame) *hash.Event {
	var (
		memberN int
		ok      bool
	)
	if memberN, ok = ss.nodes[b]; !ok {
		return nil
	}

	if f < 1 {
		panic("sanity check: Frame idx starts from 1")
	}

	event1 := ss.events[a]
	for _, event2 := range ss.roots[memberN][f] {
		if ss.sufficientCoherence(event1, event2) {
			res := event2.Hash()
			return &res
		}
	}

	return nil
}

func (ss *Strongly) Add(e *inter.Event, f idx.Frame) {
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

	if f != 0 {
		ss.roots[event.MemberN][f] = append(ss.roots[event.MemberN][f], event)
	}
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
	e.LowestSees[e.MemberN] = idx.Event(e.Index) // TODO: change e.Index type to idx.Event
	e.HighestSeen[e.MemberN] = idx.Event(e.Index)

	for p := range e.Parents {
		if p.IsZero() {
			continue
		}
		parent := ss.events[p]
		ss.updateAllLowestSees(parent, e.MemberN, idx.Event(e.Index))
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
		e.LowestSees[node] > ref ||
		(node == e.MemberN && e.LowestSees[node] <= idx.Event(e.Index)) { // TODO: change e.Index type to idx.Event
		e.LowestSees[node] = ref
		return true
	}
	return false
}

// sufficientCoherence calculates "sufficient coherence" between the events.
// The event1.HighestSeen array remembers the sequence number of the last
// event by each member that is an ancestor of event1. The array for
// event2.LowestSees remembers the sequence number of the earliest
// event by each member that is a descendant of event2. Compare the two arrays,
// and find how many elements in the event1.HighestSeen array are greater
// than or equal to the corresponding element of the event2.LowestSees
// array. If there are more than 2n/3 such matches, then the event1 and event2
// have achieved sufficient coherency.
func (ss *Strongly) sufficientCoherence(event1, event2 *Event) bool {
	counter := ss.members.NewCounter()

	for m := range ss.members {
		n := ss.nodes[m]
		if event2.LowestSees[n] <= event1.HighestSeen[n] && event2.LowestSees[n] != 0 {
			counter.Count(m)
		}
	}

	return counter.HasMajority()
}
