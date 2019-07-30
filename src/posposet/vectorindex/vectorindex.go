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
	members    internal.Members
	memberIdxs map[hash.Peer]idx.Member
	events     map[hash.Event]*Event

	logger.Instance
}

// New creates Vindex instance.
func New(members internal.Members) *Vindex {
	vi := &Vindex{
		members:    members,
		memberIdxs: members.Idxs(),
		Instance:   logger.MakeInstance(),
	}
	vi.Reset()

	return vi
}

// Reset resets buffers.
func (vi *Vindex) Reset() {
	vi.events = make(map[hash.Event]*Event)
}

// Calculate vector clocks for the event.
func (vi *Vindex) Add(e *inter.Event) {
	// sanity check
	if _, ok := vi.events[e.Hash()]; ok {
		vi.Fatalf("event %s already exists", e.Hash().String())
		return
	}

	event := &Event{
		Event:     e,
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
	e.HighestBefore[e.MemberIdx].ClaimedTime = e.LamportTime // TODO .ClaimedTime

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
