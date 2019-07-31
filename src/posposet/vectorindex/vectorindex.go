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
	events     map[hash.Event]*Event // TODO should be DB + cache
	tempEvents map[hash.Event]*Event // RAM only, used for not connected events

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
	vi.tempEvents = make(map[hash.Event]*Event)
}

// Calculate vector clocks for the event.
func (vi *Vindex) calcVectors(e *inter.Event) *Event {
	event := &Event{
		Event:      e,
		CreatorIdx: vi.memberIdxs[e.Creator],
	}

	vi.fillEventVectors(event)
	return event
}

// Calculate vector clocks for the event and save into DB.
func (vi *Vindex) Add(e *inter.Event) {
	// sanity check
	if vi.GetEvent(e.Hash()) != nil {
		vi.Fatalf("event %s already exists", e.Hash().String())
	}
	event := vi.calcVectors(e)
	vi.events[e.Hash()] = event
}

// Calculate vector clocks for the event and save into RAM.
func (vi *Vindex) AddAsTemporary(e *inter.Event) {
	event := vi.calcVectors(e)
	vi.tempEvents[e.Hash()] = event
}

func (vi *Vindex) CopyTemporaryToDb(id hash.Event) {
	vi.events[id] = vi.tempEvents[id]
}

func (vi *Vindex) EraseTemporary(id hash.Event) {
	delete(vi.tempEvents, id)
}

func (vi *Vindex) GetEvent(id hash.Event) *Event {
	event := vi.events[id]
	if event == nil {
		return vi.tempEvents[id]
	}
	return event
}

func (vi *Vindex) fillEventVectors(e *Event) {
	e.LowestAfter = make([]LowestAfter, len(vi.memberIdxs))
	e.HighestBefore = make([]HighestBefore, len(vi.memberIdxs))

	// seen by himself
	e.LowestAfter[e.CreatorIdx].Seq = e.Seq
	e.HighestBefore[e.CreatorIdx].Seq = e.Seq
	e.HighestBefore[e.CreatorIdx].Id = e.Hash()
	e.HighestBefore[e.CreatorIdx].ClaimedTime = e.ClaimedTime // TODO .ClaimedTime

	// pre-load parents into RAM for quick access
	eParents := make([]*Event, 0, len(e.Parents))
	for _, p := range e.Parents {
		eParents = append(eParents, vi.GetEvent(p))
	}

	for _, p := range eParents {
		// calculate HighestBefore vector. Detect forks for a case when parent does see a fork
		for i, high := range p.HighestBefore {
			if e.HighestBefore[i].IsForkSeen {
				continue
			}
			if high.IsForkSeen || e.HighestBefore[i].Seq < high.Seq {
				e.HighestBefore[i] = high
			}
		}
	}

	for _, p := range eParents {
		// we could just pass e.Hash() instead of the outer, but e isn't written yet
		err := vi.dfsSubgraph(p.Hash(), func(walk *Event) bool {
			if walk.LowestAfter[e.CreatorIdx].Seq != 0 {
				return false
			}
			// 'walk' is first time seen by e.Creator

			// Detect forks for a case when fork is seen only seen if we combine parents
			for _, p := range eParents {
				// p sees events older than 'walk', but p doesn't see p
				if p.HighestBefore[walk.CreatorIdx].Seq >= walk.Seq && walk.LowestAfter[p.CreatorIdx].Seq == 0 {
					e.HighestBefore[walk.CreatorIdx].IsForkSeen = true
					e.HighestBefore[walk.CreatorIdx].Seq = 0
				}
			}

			// calculate LowestAfter vector
			walk.LowestAfter[e.CreatorIdx].Seq = e.Seq
			return true
		})

		if err != nil {
			vi.Fatalf("Vindex: error during dfxSubgraph %v", err)
		}
	}
}
