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

/*
Benchmark Index.Add before optimization:
goos: linux
goarch: amd64
pkg: github.com/Fantom-foundation/go-lachesis/vector
BenchmarkIndex_Add-8   	  118716	      9639 ns/op
PASS
ok  	github.com/Fantom-foundation/go-lachesis/vector	4.318s

*/

const (
	forklessCauseCacheSize = 5000
	highestBeforeSeqCacheSize = 100
	highestBeforeTimeCacheSize = 100
	lowestAfterSeqCacheSize = 100
	eventBranchCacheSize = 100
	branchesInfoCacheSize = 100
)

// Index is a data to detect forkless-cause condition, calculate median timestamp, detect forks.
type Index struct {
	validators    pos.Validators
	validatorIdxs map[common.Address]idx.Validator

	bi *branchesInfo

	getEvent func(hash.Event) *inter.EventHeaderData

	vecDb kvdb.FlushableKeyValueStore
	table struct {
		HighestBeforeSeq  kvdb.KeyValueStore `table:"S"`
		HighestBeforeTime kvdb.KeyValueStore `table:"T"`
		LowestAfterSeq    kvdb.KeyValueStore `table:"s"`

		EventBranch  kvdb.KeyValueStore `table:"b"`
		BranchesInfo kvdb.KeyValueStore `table:"B"`
	}

	cache struct {
		HighestBeforeSeq  *lru.Cache
		HighestBeforeTime *lru.Cache
		LowestAfterSeq    *lru.Cache

		EventBranch  *lru.Cache
		BranchesInfo *lru.Cache
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
	vi.cache.HighestBeforeSeq, _ = lru.New(highestBeforeSeqCacheSize)
	vi.cache.HighestBeforeTime, _ = lru.New(highestBeforeTimeCacheSize)
	vi.cache.LowestAfterSeq, _ = lru.New(lowestAfterSeqCacheSize)
	vi.cache.BranchesInfo, _ = lru.New(branchesInfoCacheSize)
	vi.cache.EventBranch, _ = lru.New(eventBranchCacheSize)
	vi.Reset(validators, db, getEvent)

	return vi
}

// Reset resets buffers.
func (vi *Index) Reset(validators pos.Validators, db kvdb.KeyValueStore, getEvent func(hash.Event) *inter.EventHeaderData) {
	// we use wrapper to be able to drop failed events by dropping cache
	vi.getEvent = getEvent
	vi.vecDb = flushable.Wrap(db)
	vi.validators = validators.Copy()
	vi.validatorIdxs = validators.Idxs()
	vi.DropNotFlushed()
	vi.forklessCauseCache.Purge()
	vi.cleanCaches()

	table.MigrateTables(&vi.table, vi.vecDb)
}

func (vi *Index) cleanCaches() {
	vi.cache.HighestBeforeSeq.Purge()
	vi.cache.HighestBeforeTime.Purge()
	vi.cache.LowestAfterSeq.Purge()
	vi.cache.BranchesInfo.Purge()
	vi.cache.EventBranch.Purge()
}

// Add calculates vector clocks for the event and saves into DB.
func (vi *Index) Add(e *inter.EventHeaderData) {
	// sanity check
	if vi.GetHighestBeforeSeq(e.Hash()) != nil {
		vi.Log.Warn("Event already exists", "event", e.Hash().String())
		return
	}
	vi.initBranchesInfo()
	_ = vi.fillEventVectors(e)
}

// Flush writes vector clocks to persistent store.
func (vi *Index) Flush() {
	if vi.bi != nil {
		vi.setBranchesInfo(vi.bi)
	}
	if err := vi.vecDb.Flush(); err != nil {
		vi.Log.Crit("Failed to flush db", "err", err)
	}
	vi.cleanCaches()
}

// DropNotFlushed not connected clocks. Call it if event has failed.
func (vi *Index) DropNotFlushed() {
	vi.bi = nil
	vi.vecDb.DropNotFlushed()
	vi.cleanCaches()
}

func (vi *Index) fillGlobalBranchID(e *inter.EventHeaderData, meIdx idx.Validator) idx.Validator {
	// sanity checks
	if len(vi.bi.BranchIDCreators) != len(vi.bi.BranchIDLastSeq) {
		vi.Log.Crit("Inconsistent BranchIDCreators len (inconsistent DB)", "event", e.String())
	}
	if len(vi.bi.BranchIDCreators) != len(vi.bi.BranchIDCreatorIdxs) {
		vi.Log.Crit("Inconsistent BranchIDCreators len (inconsistent DB)", "event", e.String())
	}
	if len(vi.bi.BranchIDCreators) < len(vi.validators) {
		vi.Log.Crit("Inconsistent BranchIDCreators len (inconsistent DB)", "event", e.String())
	}

	if e.SelfParent() == nil {
		// is it first event indeed?
		if vi.bi.BranchIDLastSeq[meIdx] == 0 {
			// OK, not a new fork
			vi.bi.BranchIDLastSeq[meIdx] = e.Seq
			return meIdx
		}
	} else {
		selfParentBranchID := vi.getEventBranchID(*e.SelfParent())
		// sanity checks
		if len(vi.bi.BranchIDCreators) != len(vi.bi.BranchIDLastSeq) {
			vi.Log.Crit("Inconsistent BranchIDCreators len (inconsistent DB)", "event", e.String())
		}
		if len(vi.bi.BranchIDCreators) != len(vi.bi.BranchIDCreatorIdxs) {
			vi.Log.Crit("Inconsistent BranchIDCreators len (inconsistent DB)", "event", e.String())
		}
		if vi.bi.BranchIDCreators[selfParentBranchID] != e.Creator {
			vi.Log.Crit("Inconsistent BranchIDCreators (inconsistent DB). Wrong self-parent?", "event", e.String())
		}

		if vi.bi.BranchIDLastSeq[selfParentBranchID]+1 == e.Seq {
			vi.bi.BranchIDLastSeq[selfParentBranchID] = e.Seq
			// OK, not a new fork
			return selfParentBranchID
		}
	}

	// if we're here, then new fork is observed (only globally), create new branchID due to a new fork
	vi.bi.BranchIDLastSeq = append(vi.bi.BranchIDLastSeq, e.Seq)
	vi.bi.BranchIDCreators = append(vi.bi.BranchIDCreators, e.Creator)
	vi.bi.BranchIDCreatorIdxs = append(vi.bi.BranchIDCreatorIdxs, meIdx)
	newBranchID := idx.Validator(len(vi.bi.BranchIDLastSeq) - 1)
	vi.bi.BranchIDByCreators[meIdx] = append(vi.bi.BranchIDByCreators[meIdx], newBranchID)
	return newBranchID
}

func (vi *Index) setForkDetected(beforeSeq HighestBeforeSeq, branchID idx.Validator) {
	creatorIdx := vi.bi.BranchIDCreatorIdxs[branchID]
	for _, branchID := range vi.bi.BranchIDByCreators[creatorIdx] {
		beforeSeq.Set(idx.Validator(branchID), forkDetectedSeq)
	}
	// sanity check
	if !vi.atLeastOneFork() {
		vi.Log.Crit("Not written the correct branches info (inconsistent DB)")
	}
}

// fillEventVectors calculates (and stores) event's vectors, and updates LowestAfter of newly-observed events.
func (vi *Index) fillEventVectors(e *inter.EventHeaderData) allVecs {
	meIdx := vi.validatorIdxs[e.Creator]
	myVecs := allVecs{
		beforeSeq:  NewHighestBeforeSeq(len(vi.bi.BranchIDCreators)),
		beforeTime: NewHighestBeforeTime(len(vi.bi.BranchIDCreators)),
		after:      NewLowestAfterSeq(len(vi.bi.BranchIDCreators)),
	}

	meBranchID := vi.fillGlobalBranchID(e, meIdx)

	// pre-load parents into RAM for quick access
	parentsVecs := make([]allVecs, len(e.Parents))
	parentsBranchIDs := make([]idx.Validator, len(e.Parents))
	for i, p := range e.Parents {
		parentsBranchIDs[i] = vi.getEventBranchID(p)
		parentsVecs[i] = allVecs{
			beforeSeq:  vi.GetHighestBeforeSeq(p),
			beforeTime: vi.GetHighestBeforeTime(p),
			//after : vi.GetLowestAfterSeq(p), not needed
		}
		if parentsVecs[i].beforeSeq == nil {
			vi.Log.Crit("Processed out of order, parent not found (inconsistent DB)", "parent", p.String())
		}
	}

	// observed by himself
	myVecs.after.Set(meBranchID, e.Seq)
	myVecs.beforeSeq.Set(meBranchID, BranchSeq{Seq: e.Seq, MinSeq: e.Seq})
	myVecs.beforeTime.Set(meBranchID, e.ClaimedTime)

	for _, pVec := range parentsVecs {
		// calculate HighestBefore vector. Detect forks for a case when parent observes a fork
		for branchID := idx.Validator(0); branchID < idx.Validator(len(vi.bi.BranchIDCreators)); branchID++ {
			hisSeq := pVec.beforeSeq.Get(branchID)
			if hisSeq.Seq == 0 && !hisSeq.IsForkDetected() {
				// hisSeq doesn't observe anything about this branchID
				continue
			}
			mySeq := myVecs.beforeSeq.Get(branchID)

			if mySeq.IsForkDetected() {
				// mySeq observes the maximum already
				continue
			}
			if hisSeq.IsForkDetected() {
				// set fork detected
				vi.setForkDetected(myVecs.beforeSeq, branchID)
			} else {
				if mySeq.Seq == 0 || mySeq.MinSeq > hisSeq.MinSeq {
					// take hisSeq.MinSeq
					mySeq.MinSeq = hisSeq.MinSeq
					myVecs.beforeSeq.Set(branchID, mySeq)
				}
				if mySeq.Seq < hisSeq.Seq {
					// take hisSeq.Seq
					mySeq.Seq = hisSeq.Seq
					myVecs.beforeSeq.Set(branchID, mySeq)
					myVecs.beforeTime.Set(branchID, pVec.beforeTime.Get(branchID))
				}
			}
		}
	}
	// Detect forks, which were not observed by parents
	for n := idx.Validator(0); n < idx.Validator(len(vi.validators)); n++ {
		if myVecs.beforeSeq.Get(n).IsForkDetected() {
			// fork is already detected from the creator
			continue
		}
		for _, branchID1 := range vi.bi.BranchIDByCreators[n] {
			for _, branchID2 := range vi.bi.BranchIDByCreators[n] {
				if branchID1 == branchID2 {
					continue
				}

				a := myVecs.beforeSeq.Get(branchID1)
				b := myVecs.beforeSeq.Get(branchID2)

				if a.Seq == 0 || b.Seq == 0 {
					continue
				}
				if a.MinSeq <= b.Seq && b.MinSeq <= a.Seq {
					vi.setForkDetected(myVecs.beforeSeq, n)
					goto nextCreator
				}
			}
		}
	nextCreator:
	}

	for _, headP := range e.Parents {
		// we could just pass e.Hash() instead of the outer loop, but e isn't written yet
		walk := func(w *inter.EventHeaderData) (godeeper bool) {
			wLowestAfterSeq := vi.GetLowestAfterSeq(w.Hash())

			godeeper = wLowestAfterSeq.Get(meBranchID) == 0
			if !godeeper {
				return
			}

			// update LowestAfter vector of the old event, because newly-connected event observes it
			wLowestAfterSeq.Set(meBranchID, e.Seq)
			vi.SetLowestAfter(w.Hash(), wLowestAfterSeq)

			return
		}

		err := vi.dfsSubgraph(headP, walk)
		if err != nil {
			vi.Log.Crit("VectorClock: Failed to walk subgraph", "err", err)
		}
	}

	// store calculated vectors
	vi.SetHighestBefore(e.Hash(), myVecs.beforeSeq, myVecs.beforeTime)
	vi.SetLowestAfter(e.Hash(), myVecs.after)
	vi.setEventBranchID(e.Hash(), meBranchID)

	return myVecs
}

// GetHighestBeforeAllBranches returns HighestBefore vector clock without branches, where branches are merged into one
func (vi *Index) GetHighestBeforeAllBranches(id hash.Event) HighestBeforeSeq {
	mergedSeq, _ := vi.getHighestBeforeAllBranchesTime(id)
	return mergedSeq
}

func (vi *Index) getHighestBeforeAllBranchesTime(id hash.Event) (HighestBeforeSeq, HighestBeforeTime) {
	vi.initBranchesInfo()

	if vi.atLeastOneFork() {
		beforeSeq := vi.GetHighestBeforeSeq(id)
		times := vi.GetHighestBeforeTime(id)
		mergedTimes := NewHighestBeforeTime(len(vi.validators))
		mergedSeq := NewHighestBeforeSeq(len(vi.validators))
		for creatorIdx, branches := range vi.bi.BranchIDByCreators {
			// read all branches to find highest event
			highestBranchSeq := BranchSeq{}
			highestBranchTime := inter.Timestamp(0)
			for _, branchID := range branches {
				branch := beforeSeq.Get(branchID)
				if branch.IsForkDetected() {
					highestBranchSeq = branch
					break
				}
				if branch.Seq > highestBranchSeq.Seq {
					highestBranchSeq = branch
					highestBranchTime = times.Get(branchID)
				}
			}
			mergedTimes.Set(idx.Validator(creatorIdx), highestBranchTime)
			mergedSeq.Set(idx.Validator(creatorIdx), highestBranchSeq)
		}

		return mergedSeq, mergedTimes
	}
	return vi.GetHighestBeforeSeq(id), vi.GetHighestBeforeTime(id)
}
