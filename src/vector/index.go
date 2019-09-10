package vector

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

const (
	stronglySeeCacheSize = 5000
)

// Index is a data to detect strongly-see condition, calculate median timestamp, detect forks.
type Index struct {
	members    pos.Members
	memberIdxs map[common.Address]idx.Member
	getEvent   func(hash.Event) *inter.EventHeaderData

	vecDb kvdb.FlushableKeyValueStore
	table struct {
		HighestBeforeSeq  kvdb.KeyValueStore `table:"S"`
		HighestBeforeTime kvdb.KeyValueStore `table:"T"`
		LowestAfterSeq    kvdb.KeyValueStore `table:"s"`
	}

	stronglySeeCache *lru.Cache

	logger.Instance
}

// NewIndex creates Index instance.
func NewIndex(members pos.Members, db kvdb.KeyValueStore, getEvent func(hash.Event) *inter.EventHeaderData) *Index {
	cache, _ := lru.New(stronglySeeCacheSize)

	vi := &Index{
		Instance:         logger.MakeInstance(),
		stronglySeeCache: cache,
	}
	vi.Reset(members, db, getEvent)

	return vi
}

// Reset resets buffers.
func (vi *Index) Reset(members pos.Members, db kvdb.KeyValueStore, getEvent func(hash.Event) *inter.EventHeaderData) {
	// we use wrapper to be able to drop failed events by dropping cache
	vi.getEvent = getEvent
	vi.vecDb = flushable.New(db)
	vi.members = members
	vi.memberIdxs = members.Idxs()

	table.MigrateTables(&vi.table, vi.vecDb)
}

// Add calculates vector clocks for the event and saves into DB.
func (vi *Index) Add(e *inter.EventHeaderData) {
	// sanity check
	if vi.GetHighestBeforeSeq(e.Hash()) != nil {
		vi.Log.Crit("Event already exists", "event", e.Hash().String())
	}
	vecs := vi.fillEventVectors(e)
	vi.SetHighestBefore(e.Hash(), vecs.beforeSee, vecs.beforeTime)
	vi.SetLowestAfter(e.Hash(), vecs.afterSee)
}

// Flush writes vector clocks to persistent store.
func (vi *Index) Flush() {
	if err := vi.vecDb.Flush(); err != nil {
		vi.Log.Crit("Failed to flush db", "err", err)
	}
}

// DropNotFlushed not connected clocks. Call it if event has failed.
func (vi *Index) DropNotFlushed() {
	vi.vecDb.DropNotFlushed()
}

func (vi *Index) fillEventVectors(e *inter.EventHeaderData) allVecs {
	meIdx := vi.memberIdxs[e.Creator]
	myVecs := allVecs{
		beforeSee:  NewHighestBeforeSeq(len(vi.memberIdxs)),
		beforeTime: NewHighestBeforeTime(len(vi.memberIdxs)),
		afterSee:   NewLowestAfterSeq(len(vi.memberIdxs)),
	}

	// seen by himself
	myVecs.afterSee.Set(meIdx, e.Seq)
	myVecs.beforeSee.Set(meIdx, ForkSeq{Seq: e.Seq})
	myVecs.beforeTime.Set(meIdx, e.ClaimedTime)

	// pre-load parents into RAM for quick access
	parentsVecs := make([]allVecs, len(e.Parents))
	parentsCreators := make([]idx.Member, len(e.Parents))
	for i, p := range e.Parents {
		parent := vi.getEvent(p)
		if parent == nil {
			vi.Log.Crit("Event %s wasn't found", "event", p.String())
		}
		parentsCreators[i] = vi.memberIdxs[parent.Creator]
		parentsVecs[i] = allVecs{
			beforeSee:  vi.GetHighestBeforeSeq(p),
			beforeTime: vi.GetHighestBeforeTime(p),
			//afterSee : vi.GetLowestAfterSeq(p), not needed
		}
		if parentsVecs[i].beforeSee == nil {
			vi.Log.Crit("processed out of order, parent wasn't found", "parent", p.String())
		}
	}

	for _, pVec := range parentsVecs {
		// calculate HighestBefore vector. Detect forks for a case when parent does see a fork
		for n := idx.Member(0); n < idx.Member(len(vi.memberIdxs)); n++ {
			myForkSeq := myVecs.beforeSee.Get(n)
			hisForkSeq := pVec.beforeSee.Get(n)

			if myForkSeq.IsForkSeen {
				continue
			}
			if hisForkSeq.IsForkSeen || myForkSeq.Seq < hisForkSeq.Seq {
				myVecs.beforeSee.Set(n, hisForkSeq)
				myVecs.beforeTime.Set(n, pVec.beforeTime.Get(n))
			}
		}
	}

	for _, pVec := range parentsVecs {
		hisForkSeq := pVec.beforeSee.Get(meIdx)
		// self-fork detection
		if hisForkSeq.Seq >= e.Seq {
			myVecs.beforeSee.Set(meIdx, ForkSeq{IsForkSeen: true, Seq: 0})
		}
	}

	for _, headP := range e.Parents {
		// we could just pass e.Hash() instead of the outer loop, but e isn't written yet
		walk := func(w *inter.EventHeaderData) (godeeper bool) {
			wLowestAfterSeq := vi.GetLowestAfterSeq(w.Hash())
			godeeper = wLowestAfterSeq.Get(meIdx) == 0
			if !godeeper {
				return
			}

			wCreatorIdx := vi.memberIdxs[w.Creator]

			// 'walk' is first time seen by e.Creator
			// Detect forks for a case when fork is seen only seen if we combine parents
			for i, pVec := range parentsVecs {
				if pVec.beforeSee.Get(wCreatorIdx).Seq >= w.Seq && wLowestAfterSeq.Get(parentsCreators[i]) == 0 {
					myVecs.beforeSee.Set(wCreatorIdx, ForkSeq{IsForkSeen: true, Seq: 0})
				}
			}

			// calculate LowestAfter vector
			wLowestAfterSeq.Set(meIdx, e.Seq)
			vi.SetLowestAfter(w.Hash(), wLowestAfterSeq)

			return
		}

		err := vi.dfsSubgraph(headP, walk)
		if err != nil {
			vi.Log.Crit("Error during dfxSubgraph", "err", err)
		}
	}

	return myVecs
}
