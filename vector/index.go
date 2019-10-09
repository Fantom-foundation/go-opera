package vector

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

const (
	forklessCauseCacheSize = 5000
)

// Index is a data to detect forkless-cause condition, calculate median timestamp, detect forks.
type Index struct {
	validators    pos.Validators
	validatorIdxs map[common.Address]idx.Validator
	getEvent      func(hash.Event) *inter.EventHeaderData

	vecDb kvdb.FlushableKeyValueStore
	table struct {
		HighestBeforeSeq  kvdb.KeyValueStore `table:"S"`
		HighestBeforeTime kvdb.KeyValueStore `table:"T"`
		LowestAfterSeq    kvdb.KeyValueStore `table:"s"`
	}

	forklessCauseCache *lru.Cache

	logger.Instance
}

// NewIndex creates Index instance.
func NewIndex(validators pos.Validators, db kvdb.KeyValueStore, getEvent func(hash.Event) *inter.EventHeaderData) *Index {
	cache, _ := lru.New(forklessCauseCacheSize)

	vi := &Index{
		Instance:           logger.MakeInstance(),
		forklessCauseCache: cache,
	}
	vi.Reset(validators, db, getEvent)

	return vi
}

// Reset resets buffers.
func (vi *Index) Reset(validators pos.Validators, db kvdb.KeyValueStore, getEvent func(hash.Event) *inter.EventHeaderData) {
	// we use wrapper to be able to drop failed events by dropping cache
	vi.getEvent = getEvent
	vi.vecDb = flushable.Wrap(db)
	vi.validators = validators
	vi.validatorIdxs = validators.Idxs()

	table.MigrateTables(&vi.table, vi.vecDb)
}

// Add calculates vector clocks for the event and saves into DB.
func (vi *Index) Add(e *inter.EventHeaderData) {
	// sanity check
	if vi.GetHighestBeforeSeq(e.Hash()) != nil {
		vi.Log.Warn("Event already exists", "event", e.Hash().String())
		return
	}
	_ = vi.fillEventVectors(e)
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

// fillEventVectors calculates (and stores) event's vectors, and updates LowestAfter of newly-observed events.
func (vi *Index) fillEventVectors(e *inter.EventHeaderData) allVecs {
	meIdx := vi.validatorIdxs[e.Creator]
	myVecs := allVecs{
		beforeCause: NewHighestBeforeSeq(len(vi.validatorIdxs)),
		beforeTime:  NewHighestBeforeTime(len(vi.validatorIdxs)),
		afterCause:  NewLowestAfterSeq(len(vi.validatorIdxs)),
	}

	// observed by himself
	myVecs.afterCause.Set(meIdx, e.Seq)
	myVecs.beforeCause.Set(meIdx, ForkSeq{Seq: e.Seq})
	myVecs.beforeTime.Set(meIdx, e.ClaimedTime)

	// pre-load parents into RAM for quick access
	parentsVecs := make([]allVecs, len(e.Parents))
	parentsCreators := make([]idx.Validator, len(e.Parents))
	for i, p := range e.Parents {
		parent := vi.getEvent(p)
		if parent == nil {
			vi.Log.Crit("Event not found", "event", p.String())
		}
		parentsCreators[i] = vi.validatorIdxs[parent.Creator]
		parentsVecs[i] = allVecs{
			beforeCause: vi.GetHighestBeforeSeq(p),
			beforeTime:  vi.GetHighestBeforeTime(p),
			//afterCause : vi.GetLowestAfterSeq(p), not needed
		}
		if parentsVecs[i].beforeCause == nil {
			vi.Log.Crit("Processed out of order, parent not found", "parent", p.String())
		}
	}

	for _, pVec := range parentsVecs {
		// calculate HighestBefore vector. Detect forks for a case when parent observes a fork
		for n := idx.Validator(0); n < idx.Validator(len(vi.validatorIdxs)); n++ {
			myForkSeq := myVecs.beforeCause.Get(n)
			hisForkSeq := pVec.beforeCause.Get(n)

			if myForkSeq.IsForkDetected {
				continue
			}
			if hisForkSeq.IsForkDetected || myForkSeq.Seq < hisForkSeq.Seq {
				myVecs.beforeCause.Set(n, hisForkSeq)
				myVecs.beforeTime.Set(n, pVec.beforeTime.Get(n))
			}
		}
	}

	for _, pVec := range parentsVecs {
		hisForkSeq := pVec.beforeCause.Get(meIdx)
		// self-fork detection
		if hisForkSeq.Seq >= e.Seq {
			myVecs.beforeCause.Set(meIdx, ForkSeq{IsForkDetected: true, Seq: 0})
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

			wCreatorIdx := vi.validatorIdxs[w.Creator]

			// 'walk' is first time observed by e.Creator
			// Detect forks for a case when fork is detected only if we combine parents
			for i, pVec := range parentsVecs {
				if pVec.beforeCause.Get(wCreatorIdx).Seq >= w.Seq && wLowestAfterSeq.Get(parentsCreators[i]) == 0 {
					myVecs.beforeCause.Set(wCreatorIdx, ForkSeq{IsForkDetected: true, Seq: 0})
				}
			}

			// update LowestAfter vector of the newly-observed event
			wLowestAfterSeq.Set(meIdx, e.Seq)
			vi.SetLowestAfter(w.Hash(), wLowestAfterSeq)

			return
		}

		err := vi.dfsSubgraph(headP, walk)
		if err != nil {
			vi.Log.Crit("VectorClock: Failed to walk subgraph", "err", err)
		}
	}

	// store calculated vectors
	vi.SetHighestBefore(e.Hash(), myVecs.beforeCause, myVecs.beforeTime)
	vi.SetLowestAfter(e.Hash(), myVecs.afterCause)

	return myVecs
}
