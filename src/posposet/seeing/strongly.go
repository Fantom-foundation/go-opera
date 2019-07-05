package seeing

import (
	"math"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
)

const membersNumber = 30

// Strongly is a datas to check strongly-see relation..
type Strongly struct {
	members internal.Members
	events  map[hash.Event]*Event

	firstDescendantsSeq []int64
	lastAncestorsSeq    []int64
	recentEvents        []hash.Event      // index is a member index.
	membersByPears      map[hash.Peer]int // mapping creator id -> member num

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
	ss.events = make(map[hash.Event]*Event)
	ss.recentEvents = make([]hash.Event, len(mm))
	ss.firstDescendantsSeq = make([]int64, len(mm))
	ss.lastAncestorsSeq = make([]int64, len(mm))

	for i := 0; i < len(mm); i++ {
		ss.firstDescendantsSeq[i] = math.MaxInt32
		ss.lastAncestorsSeq[i] = -1
	}
}

func (ss *Strongly) Add(e *inter.Event) {
	// sanity check
	if _, ok := ss.events[e.Hash()]; ok {
		ss.Fatalf("event %s already exists", e.Hash().String())
		return
	}

	event := &Event{
		Event: e,
	}
	ss.events[e.Hash()] = event
	ss.fillEventSequences(event)
	ss.fillFirstAncestors(event)
}

func (ss *Strongly) Seen(a, b hash.Event) bool {
	a1 := ss.events[a]
	b1 := ss.events[b]

	return ss.sufficientCoherence(a1, b1)
}

func (ss *Strongly) findMemberNumber(creator hash.Peer) (int, bool) {
	num, ok := ss.membersByPears[creator]
	if !ok {
		return 0, false
	}
	return num, true
}

func (ss *Strongly) highestEventFromMember(member int) (hash.Event, error) {
	if member >= len(ss.recentEvents) {
		return hash.Event{}, ErrInvalidMemberNum
	}
	return ss.recentEvents[member], nil
}

func (ss *Strongly) memberLastAncestorSeq(member int) int64 {
	if member >= len(ss.lastAncestorsSeq) {
		return -1 // default value to last ancestor sequence.
	}
	return ss.lastAncestorsSeq[member]
}

// sufficientCoherence calculates "sufficient coherence" between the events.
// The event1.lastAncestorsSeq array remembers the sequence number of the last
// event by each member that is an ancestor of event1. The array for
// event2.FirstDescendantsSeq remembers the sequence number of the earliest
// event by each member that is a descendant of event2. Compare the two arrays,
// and find how many elements in the event1.lastAncestorsSeq array are greater
// than or equal to the corresponding element of the event2.FirstDescendantsSeq
// array. If there are more than 2n/3 such matches, then the event1 and event2
// have achieved sufficient coherency.
func (ss *Strongly) sufficientCoherence(event1, event2 *Event) bool {
	if len(event1.LastAncestorsSeq) != len(event2.FirstDescendantsSeq) {
		return false
	}

	counter := 0
	for k := range event1.LastAncestorsSeq {
		if event2.FirstDescendantsSeq[k] <= event1.LastAncestorsSeq[k] {
			counter++
		}
	}

	if counter >= len(event1.LastAncestorsSeq)*2/3 {
		return true
	}

	return false
}

func (ss *Strongly) fillEventSequences(event *Event) {
	memberNumber, ok := ss.findMemberNumber(event.Creator)
	if !ok {
		return
	}

	var (
		foundSelfParent  bool
		foundOtherParent bool
	)

	getOtherParent := func() *Event {
		// TODO: we need to determine the number of other parents in the future.
		op := event.OtherParents()[0] // take a first other parent.
		return ss.events[op]
	}

	initLastAncestors := func() {
		if len(event.LastAncestors) == len(ss.members) {
			return
		}
		event.LastAncestors = make([]hash.Event, len(ss.members))
	}

	selfParent, found := event.SelfParent()
	if found {
		foundSelfParent = true
	}

	otherParents := event.OtherParents()
	if otherParents.Len() > 0 {
		foundOtherParent = true
	}

	if !foundSelfParent && !foundOtherParent {
		event.LastAncestorsSeq = ss.lastAncestorsSeq

		highestEvent, err := ss.highestEventFromMember(memberNumber)
		if err != nil {
			ss.Fatal(err.Error())
			return
		}

		initLastAncestors()
		event.LastAncestors[memberNumber] = highestEvent
	} else if !foundSelfParent {
		parent := getOtherParent()
		event.LastAncestors = parent.LastAncestors
		event.LastAncestorsSeq = parent.LastAncestorsSeq
	} else if !foundOtherParent {
		parent := ss.events[selfParent]
		event.LastAncestors = parent.LastAncestors
		event.LastAncestorsSeq = parent.LastAncestorsSeq
	} else {
		sp := ss.events[selfParent]
		event.LastAncestors = sp.LastAncestors
		event.LastAncestorsSeq = sp.LastAncestorsSeq

		otherParent := getOtherParent()

		for i := 0; i < len(ss.members); i++ {
			if event.LastAncestorsSeq[i] >= otherParent.LastAncestorsSeq[i] {
				event.LastAncestors[i] = otherParent.LastAncestors[i]
				event.LastAncestorsSeq[i] = otherParent.LastAncestorsSeq[i]
			}
		}
	}

	event.FirstDescendantsSeq = ss.firstDescendantsSeq
	event.FirstDescendants = ss.recentEvents

	event.LastAncestors[memberNumber] = event.Hash()
	event.FirstDescendants[memberNumber] = event.Hash()

	event.FirstDescendantsSeq[memberNumber] =
		ss.memberLastAncestorSeq(memberNumber)
	event.LastAncestorsSeq[memberNumber] =
		ss.memberLastAncestorSeq(memberNumber)
}

func (ss *Strongly) fillFirstAncestors(event *Event) {
	memberNumber, ok := ss.findMemberNumber(event.Creator)
	if !ok {
		return
	}

	currentSeq := ss.lastAncestorsSeq[memberNumber]

	for _, h := range event.LastAncestors {
		lastAncestor := ss.events[h]

		for lastAncestor != nil &&
			!lastAncestor.FirstDescendants[memberNumber].IsZero() {
			lastAncestor.FirstDescendantsSeq[memberNumber] = currentSeq
			lastAncestor.FirstDescendants[memberNumber] = event.Hash()

			parent, found := lastAncestor.SelfParent()
			if !found {
				break
			}

			if p, ok := ss.events[parent]; ok {
				lastAncestor = p
				continue
			}
			break
		}
	}

}
