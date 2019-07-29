package vectorindex

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
)

// Vindex is a data to detect strongly-see condition, calculate median timestamp, detect forks.
type Vindex struct {
	newCounter internal.StakeCounterProvider
	memberIdxs  map[hash.Peer]idx.Member
	events     map[hash.Event]*Event

	logger.Instance
}

// New creates Vindex instance.
func New(c internal.StakeCounterProvider, memberIdx map[hash.Peer]idx.Member) *Vindex {
	vi := &Vindex{
		newCounter: c,
		Instance:   logger.MakeInstance(),
	}
	vi.Reset(memberIdx)

	return vi
}

// Reset resets buffers.
func (vi *Vindex) Reset(memberIdx map[hash.Peer]idx.Member) {
	vi.memberIdxs = memberIdx
	vi.events = make(map[hash.Event]*Event)
}

// Cache event for Vindex.See.
func (vi *Vindex) Cache(e *inter.Event) {
	// sanity check
	if _, ok := vi.events[e.Hash()]; ok {
		vi.Fatalf("event %s already exists", e.Hash().String())
		return
	}

	event := &Event{
		Event:   e,
		MemberIdx: vi.memberIdxs[e.Creator],
	}

	vi.fillEventRefs(event)
	vi.events[e.Hash()] = event
}

func (vi *Vindex) fillEventRefs(e *Event) {
	e.LowestAfter = make([]LowestAfter, len(vi.memberIdxs))
	e.HighestBefore = make([]HighestBefore, len(vi.memberIdxs))

	// seen by himself
	e.LowestAfter[e.MemberIdx].Seq = e.Seq
	e.HighestBefore[e.MemberIdx].Seq = e.Seq
	e.HighestBefore[e.MemberIdx].Id = e.Hash()
	//e.HighestBefore[e.MemberIdx].ClaimedTime = e.ClaimedTime

	for p := range e.Parents {
		if p.IsZero() {
			continue
		}
		parent := vi.events[p]
		vi.updateAllLowestAfter(parent, e.MemberIdx, e.Seq)
		vi.updateAllHighestBefore(e, parent)
	}
}

func (vi *Vindex) updateAllHighestBefore(e, parent *Event) {
	for i, n := range parent.HighestBefore {
		if e.HighestBefore[i].Seq < n.Seq {
			e.HighestBefore[i] = n
		}
	}
}

func (vi *Vindex) updateAllLowestAfter(e *Event, member idx.Member, ref idx.Event) {
	toUpdate := []*Event{e}
	for {
		var next []*Event
		for _, event := range toUpdate {
			if !setLowestAfterIfMin(event, member, ref) {
				continue
			}
			for p := range event.Parents {
				if !p.IsZero() {
					next = append(next, vi.events[p])
				}
			}
		}

		if len(next) == 0 {
			break
		}
		toUpdate = next
	}
}

func setLowestAfterIfMin(e *Event, member idx.Member, ref idx.Event) bool {
	curr := e.LowestAfter[member].Seq
	if curr == 0 || curr > ref {
		e.LowestAfter[member].Seq = ref
		return true
	}
	return false
}

// StronglySee calculates "sufficient coherence" between the events.
// The A.HighestBefore array remembers the sequence number of the last
// event by each member that is an ancestor of A. The array for
// B.LowestAfter remembers the sequence number of the earliest
// event by each member that is a descendant of B. Compare the two arrays,
// and find how many elements in the A.HighestBefore array are greater
// than or equal to the corresponding element of the B.LowestAfter
// array. If there are more than 2n/3 such matches, then the A and B
// have achieved sufficient coherency.
func (vi *Vindex) StronglySee(aHash, bHash hash.Event) bool {
	// get events by hash
	a, ok := vi.events[aHash]
	if !ok {
		return false
	}
	b, ok := vi.events[bHash]
	if !ok {
		return false
	}

	yes := vi.newCounter()
	no := vi.newCounter()

	// calculate strongly seeing using the indexes
	for m, n := range vi.memberIdxs {
		bLowestAfter := b.LowestAfter[n]
		aHighestBefore := a.HighestBefore[n]
		if bLowestAfter.Seq <= aHighestBefore.Seq && bLowestAfter.Seq != 0 {
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
