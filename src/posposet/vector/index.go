package vector

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
)

// Index is a data to detect strongly-see condition, calculate median timestamp, detect forks.
type Index struct {
	members    internal.Members
	memberIdxs map[hash.Peer]idx.Member
	eventsDb   kvdb.FlushableDatabase

	logger.Instance
}

// NewIndex creates Index instance.
func NewIndex(members internal.Members, db kvdb.Database) *Index {
	vi := &Index{
		Instance: logger.MakeInstance(),
	}
	vi.Reset(members, db)

	return vi
}

// Reset resets buffers.
func (vi *Index) Reset(members internal.Members, db kvdb.Database) {
	// we use wrapper to be able to drop failed events by dropping cache
	vi.eventsDb = kvdb.NewCacheWrapper(db)
	vi.members = members
	vi.memberIdxs = members.Idxs()
}

// Add calculates vector clocks for the event and saves into DB.
func (vi *Index) Add(e *inter.Event) {
	// sanity check
	if vi.GetEvent(e.Hash()) != nil {
		vi.Fatalf("event %s already exists", e.Hash().String())
	}
	w := vi.wrapEvent(e)
	vi.SetEvent(w)
}

func (vi *Index) wrapEvent(e *inter.Event) *event {
	w := &event{
		EventHeaderData: &e.EventHeaderData,
		CreatorIdx:      vi.memberIdxs[e.Creator],
	}

	vi.fillEventVectors(w)
	return w
}

// Flush writes vector clocks to persistent store.
func (vi *Index) Flush() {
	if err := vi.eventsDb.Flush(); err != nil {
		vi.Fatal(err)
	}
}

// DropNotFlushed not connected clocks. Call it if event has failed.
func (vi *Index) DropNotFlushed() {
	vi.eventsDb.ClearNotFlushed()
}

func (vi *Index) fillEventVectors(e *event) {
	e.LowestAfter = make([]lowestAfter, len(vi.memberIdxs))
	e.HighestBefore = make([]highestBefore, len(vi.memberIdxs))

	// seen by himself
	e.LowestAfter[e.CreatorIdx].Seq = e.Seq
	e.HighestBefore[e.CreatorIdx].Seq = e.Seq
	e.HighestBefore[e.CreatorIdx].ID = e.Hash()
	e.HighestBefore[e.CreatorIdx].ClaimedTime = e.ClaimedTime // TODO .ClaimedTime

	// pre-load parents into RAM for quick access
	eParents := make([]*event, 0, len(e.Parents))
	for _, p := range e.Parents {
		parent := vi.GetEvent(p)
		if parent == nil {
			vi.Fatalf("vindex: event %s wasn't found", p.String())
		}
		eParents = append(eParents, parent)
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
		walk := func(w *event) (godeeper bool) {
			godeeper = w.LowestAfter[e.CreatorIdx].Seq == 0
			if !godeeper {
				return
			}
			// 'walk' is first time seen by e.Creator

			// Detect forks for a case when fork is seen only seen if we combine parents
			for _, p := range eParents {
				// p sees events older than 'walk', but p doesn't see p
				if p.HighestBefore[w.CreatorIdx].Seq >= w.Seq && w.LowestAfter[p.CreatorIdx].Seq == 0 {
					e.HighestBefore[w.CreatorIdx].IsForkSeen = true
					e.HighestBefore[w.CreatorIdx].Seq = 0
				}
			}

			// calculate LowestAfter vector
			w.LowestAfter[e.CreatorIdx].Seq = e.Seq
			vi.SetEvent(w)

			return
		}

		err := vi.dfsSubgraph(p.Hash(), walk)
		if err != nil {
			vi.Fatalf("vector.Index: error during dfxSubgraph %v", err)
		}
	}
}
