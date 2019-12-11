package topicsdb

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
)

func (tt *TopicsDb) fetchAsync(cc ...Condition) (res []*types.Log, err error) {
	if len(cc) > MaxCount {
		err = ErrTooManyTopics
		return
	}

	conditions := uint8(len(cc))
	recs := make(map[ID]*logrecBuilder)
	for _, cond := range cc {
		it := tt.table.Topic.NewIteratorWithPrefix(cond[:])
		for it.Next() {
			id := extractLogrecID(it.Key())
			topicCount := bytesToPos(it.Value())
			rec := recs[id]
			if rec == nil {
				rec = newLogrecBuilder(id, conditions, topicCount)
				recs[id] = rec
				rec.StartFetch(tt.table.Other, tt.table.Logrec)
				defer rec.StopFetch()
			}
			rec.MatchedWith(cond)
		}

		err = it.Error()
		if err != nil {
			return
		}

		it.Release()
	}

	for _, rec := range recs {
		if !rec.IsMatched() {
			continue
		}

		var r *types.Log
		r, err = rec.Build()
		if err != nil {
			return
		}
		if r != nil {
			res = append(res, r)
		}
	}

	return
}

// StartFetch log record's data when all conditions are ok.
func (rec *logrecBuilder) StartFetch(
	othersTable ethdb.Iteratee,
	logrecTable ethdb.KeyValueReader,
) {
	if rec.ok != nil {
		return
	}
	rec.ok = make(chan struct{})
	rec.ready = make(chan error)

	go func() {
		defer close(rec.ready)

		_, conditionsOk := <-rec.ok
		if !conditionsOk {
			return
		}

		rec.ready <- rec.Fetch(othersTable, logrecTable)
	}()
}

// StopFetch releases resources associated with StartFetch,
// so you should call StopFetch after StartFetch.
func (rec *logrecBuilder) StopFetch() {
	if rec.ok != nil {
		close(rec.ok)
		rec.ok = nil
	}
	rec.ready = nil
}
