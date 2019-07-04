package seeing

import (
	"math"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

const membersNumber = 30

type strongly struct {
	firstDescendantsSeq []int64
	lastAncestorsSeq    []int64
	recentEvents        []hash.Event      // index is a member index.
	membersByPears      map[hash.Peer]int // mapping creator id -> member num
}

func newStronglySeeing() strongly {
	firstDescendantsSeq := make([]int64, membersNumber)
	lastAncestorsSeq := make([]int64, membersNumber)
	recentEvents := make([]hash.Event, membersNumber)

	for i := 0; i < membersNumber; i++ {
		firstDescendantsSeq[i] = math.MaxInt32
		lastAncestorsSeq[i] = -1
	}

	return strongly{
		firstDescendantsSeq: firstDescendantsSeq,
		lastAncestorsSeq:    lastAncestorsSeq,
		recentEvents:        recentEvents,
	}
}

func (ss *strongly) findMemberNumber(creator hash.Peer) (int, bool) {
	num, ok := ss.membersByPears[creator]
	if !ok {
		return 0, false
	}
	return num, true
}

func (ss *strongly) highestEventFromMember(member int) (hash.Event, error) {
	if member >= len(ss.recentEvents) {
		return hash.Event{}, ErrInvalidMemberNum
	}
	return ss.recentEvents[member], nil
}

func (ss *strongly) memberLastAncestorSeq(member int) int64 {
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
func (ss *strongly) sufficientCoherence(event1, event2 *Event) bool {
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

func (p *Poset) fillEventSequences(event *Event) {
	memberNumber, ok := p.findMemberNumber(event.Creator)
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
		return p.GetEvent(op)
	}

	initLastAncestors := func() {
		if len(event.LastAncestors) == len(p.members) {
			return
		}
		event.LastAncestors = make([]hash.Event, len(p.members))
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
		event.LastAncestorsSeq = p.lastAncestorsSeq

		highestEvent, err := p.highestEventFromMember(memberNumber)
		if err != nil {
			p.Fatal(err.Error())
			return
		}

		initLastAncestors()
		event.LastAncestors[memberNumber] = highestEvent
	} else if !foundSelfParent {
		parent := getOtherParent()
		event.LastAncestors = parent.LastAncestors
		event.LastAncestorsSeq = parent.LastAncestorsSeq
	} else if !foundOtherParent {
		parent := p.GetEvent(selfParent)
		event.LastAncestors = parent.LastAncestors
		event.LastAncestorsSeq = parent.LastAncestorsSeq
	} else {
		sp := p.GetEvent(selfParent)
		event.LastAncestors = sp.LastAncestors
		event.LastAncestorsSeq = sp.LastAncestorsSeq

		otherParent := getOtherParent()

		for i := 0; i < len(p.members); i++ {
			if event.LastAncestorsSeq[i] >= otherParent.LastAncestorsSeq[i] {
				event.LastAncestors[i] = otherParent.LastAncestors[i]
				event.LastAncestorsSeq[i] = otherParent.LastAncestorsSeq[i]
			}
		}
	}

	event.FirstDescendantsSeq = p.firstDescendantsSeq
	event.FirstDescendants = p.recentEvents

	event.LastAncestors[memberNumber] = event.Hash()
	event.FirstDescendants[memberNumber] = event.Hash()

	event.FirstDescendantsSeq[memberNumber] =
		p.memberLastAncestorSeq(memberNumber)
	event.LastAncestorsSeq[memberNumber] =
		p.memberLastAncestorSeq(memberNumber)
}

func (p *Poset) fillFirstAncestors(event *Event) {
	memberNumber, ok := p.findMemberNumber(event.Creator)
	if !ok {
		return
	}

	currentSeq := p.lastAncestorsSeq[memberNumber]

	for _, h := range event.LastAncestors {
		lastAncestor := p.GetEvent(h)

		for lastAncestor != nil &&
			!lastAncestor.FirstDescendants[memberNumber].IsZero() {
			lastAncestor.FirstDescendantsSeq[memberNumber] = currentSeq
			lastAncestor.FirstDescendants[memberNumber] = event.Hash()

			parent, found := lastAncestor.SelfParent()
			if !found {
				break
			}

			if p.HasEvent(parent) {
				lastAncestor = p.GetEvent(parent)
				continue
			}
			break
		}
	}

}
