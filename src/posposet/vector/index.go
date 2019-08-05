package vector

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
)

// Index is a data to detect strongly-see condition, calculate median timestamp, detect forks.
type Index struct {
	members    internal.Members
	memberIdxs map[hash.Peer]idx.Member
	events     map[hash.Event]*event

	logger.Instance
}

// NewIndex creates Index instance.
func NewIndex(members internal.Members) *Index {
	vi := &Index{
		Instance: logger.MakeInstance(),
	}
	vi.Reset(members)

	return vi
}

// Reset resets buffers.
func (vi *Index) Reset(members internal.Members) {
	vi.members = members
	vi.memberIdxs = members.Idxs()
	vi.events = make(map[hash.Event]*event)
}

// Add calculates vector clocks for the event.
func (vi *Index) Add(e *inter.Event) {
	// sanity check
	if _, ok := vi.events[e.Hash()]; ok {
		vi.Fatalf("event %s already exists", e.Hash().String())
		return
	}

	event := &event{
		Event:      e,
		CreatorIdx: vi.memberIdxs[e.Creator],
	}

	vi.fillEventVectors(event)
	vi.events[e.Hash()] = event
}

func (vi *Index) fillEventVectors(e *event) {
	e.LowestAfter = make([]lowestAfter, len(vi.memberIdxs))
	e.HighestBefore = make([]highestBefore, len(vi.memberIdxs))

	// seen by himself
	e.LowestAfter[e.CreatorIdx].Seq = e.Seq
	e.HighestBefore[e.CreatorIdx].Seq = e.Seq
	e.HighestBefore[e.CreatorIdx].ID = e.Hash()
	e.HighestBefore[e.CreatorIdx].ClaimedTime = e.LamportTime // TODO .ClaimedTime

	// pre-load parents into RAM for quick access
	eParents := make([]*event, 0, len(e.Parents))
	for p := range e.Parents {
		if p.IsZero() {
			continue
		}
		eParents = append(eParents, vi.events[p])
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
		err := vi.dfsSubgraph(p.Hash(), func(walk *event) bool {
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
			vi.Fatalf("vector.Index: error during dfxSubgraph %v", err)
		}
	}
}
