package vector

import (
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
	forklessCauseCacheSize     = 5000
	highestBeforeSeqCacheSize  = 1000
	highestBeforeTimeCacheSize = 1000
	lowestAfterSeqCacheSize    = 1000
)

// IndexCacheConfig - config for cache sizes of Index
type IndexCacheConfig struct {
	ForklessCause     int `json:"forklessCause"`
	HighestBeforeSeq  int `json:"highestBeforeSeq"`
	HighestBeforeTime int `json:"highestBeforeTime"`
	LowestAfterSeq    int `json:"lowestAfterSeq"`
}

// IndexConfig - Index config (cache sizes)
type IndexConfig struct {
	Caches IndexCacheConfig `json:"cacheSizes"`
}

// Index is a data to detect forkless-cause condition, calculate median timestamp, detect forks.
type Index struct {
	validators    *pos.Validators
	validatorIdxs map[idx.StakerID]idx.Validator

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
		ForklessCause     *lru.Cache
	}

	cfg IndexConfig

	logger.Instance
}

// DefaultIndexConfig return default index config for tests
func DefaultIndexConfig() IndexConfig {
	return IndexConfig{
		Caches: IndexCacheConfig{
			ForklessCause:     forklessCauseCacheSize,
			HighestBeforeSeq:  highestBeforeSeqCacheSize,
			HighestBeforeTime: highestBeforeTimeCacheSize,
			LowestAfterSeq:    lowestAfterSeqCacheSize,
		},
	}
}

// NewIndex creates Index instance.
func NewIndex(config IndexConfig, validators *pos.Validators, db kvdb.KeyValueStore, getEvent func(hash.Event) *inter.EventHeaderData) *Index {
	vi := &Index{
		Instance: logger.MakeInstance(),
		cfg:      config,
	}
	vi.cache.ForklessCause, _ = lru.New(vi.cfg.Caches.ForklessCause)
	vi.cache.HighestBeforeSeq, _ = lru.New(vi.cfg.Caches.HighestBeforeSeq)
	vi.cache.HighestBeforeTime, _ = lru.New(vi.cfg.Caches.HighestBeforeTime)
	vi.cache.LowestAfterSeq, _ = lru.New(vi.cfg.Caches.LowestAfterSeq)
	vi.Reset(validators, db, getEvent)

	return vi
}

// Reset resets buffers.
func (vi *Index) Reset(validators *pos.Validators, db kvdb.KeyValueStore, getEvent func(hash.Event) *inter.EventHeaderData) {
	// we use wrapper to be able to drop failed events by dropping cache
	vi.getEvent = getEvent
	vi.vecDb = flushable.Wrap(db)
	vi.validators = validators.Copy()
	vi.validatorIdxs = validators.Idxs()
	vi.DropNotFlushed()
	vi.cache.ForklessCause.Purge()
	vi.dropDependentCaches()

	table.MigrateTables(&vi.table, vi.vecDb)
}

func (vi *Index) dropDependentCaches() {
	vi.cache.HighestBeforeSeq.Purge()
	vi.cache.HighestBeforeTime.Purge()
	vi.cache.LowestAfterSeq.Purge()
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
}

// DropNotFlushed not connected clocks. Call it if event has failed.
func (vi *Index) DropNotFlushed() {
	vi.bi = nil
	if vi.vecDb.NotFlushedPairs() != 0 {
		vi.vecDb.DropNotFlushed()
		vi.dropDependentCaches()
	}
}

func (vi *Index) fillGlobalBranchID(e *inter.EventHeaderData, meIdx idx.Validator) idx.Validator {
	// sanity checks
	if len(vi.bi.BranchIDCreatorIdxs) != len(vi.bi.BranchIDLastSeq) {
		vi.Log.Crit("Inconsistent BranchIDCreators len (inconsistent DB)", "event", e.String())
	}
	if len(vi.bi.BranchIDCreatorIdxs) < vi.validators.Len() {
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
		if len(vi.bi.BranchIDCreatorIdxs) != len(vi.bi.BranchIDLastSeq) {
			vi.Log.Crit("Inconsistent BranchIDCreators len (inconsistent DB)", "event", e.String())
		}

		if vi.bi.BranchIDLastSeq[selfParentBranchID]+1 == e.Seq {
			vi.bi.BranchIDLastSeq[selfParentBranchID] = e.Seq
			// OK, not a new fork
			return selfParentBranchID
		}
	}

	// if we're here, then new fork is observed (only globally), create new branchID due to a new fork
	vi.bi.BranchIDLastSeq = append(vi.bi.BranchIDLastSeq, e.Seq)
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
		beforeSeq:  NewHighestBeforeSeq(len(vi.bi.BranchIDCreatorIdxs)),
		beforeTime: NewHighestBeforeTime(len(vi.bi.BranchIDCreatorIdxs)),
		after:      NewLowestAfterSeq(len(vi.bi.BranchIDCreatorIdxs)),
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
		for branchID := idx.Validator(0); branchID < idx.Validator(len(vi.bi.BranchIDCreatorIdxs)); branchID++ {
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
	for n := idx.Validator(0); n < idx.Validator(vi.validators.Len()); n++ {
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

	// graph traversal starting from e, but excluding e
	onWalk := func(walk hash.Event) (godeeper bool) {
		wLowestAfterSeq := vi.GetLowestAfterSeq(walk)

		godeeper = wLowestAfterSeq.Get(meBranchID) == 0
		if !godeeper {
			return
		}

		// update LowestAfter vector of the old event, because newly-connected event observes it
		wLowestAfterSeq.Set(meBranchID, e.Seq)
		vi.SetLowestAfter(walk, wLowestAfterSeq)

		return
	}
	err := vi.dfsSubgraph(e, onWalk)
	if err != nil {
		vi.Log.Crit("VectorClock: Failed to walk subgraph", "err", err)
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
		mergedTimes := NewHighestBeforeTime(vi.validators.Len())
		mergedSeq := NewHighestBeforeSeq(vi.validators.Len())
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
