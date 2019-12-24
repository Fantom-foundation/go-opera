package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
)

func (tt *Index) fetchAsync(topics [][]common.Hash) (res []*types.Log, err error) {
	if len(topics) > MaxCount {
		err = ErrTooManyTopics
		return
	}

	var (
		recs      = make(map[ID]*logrecBuilder)
		condCount = uint8(len(topics))
		wildcards uint8
		prefix    [prefixSize]byte
	)
	for pos, cond := range topics {
		if len(cond) < 1 {
			wildcards++
			continue
		}
		copy(prefix[common.HashLength:], posToBytes(uint8(pos)))
		for _, alternative := range cond {
			copy(prefix[:], alternative[:])
			it := tt.table.Topic.NewIteratorWithPrefix(prefix[:])
			for it.Next() {
				id := extractLogrecID(it.Key())
				topicCount := bytesToPos(it.Value())
				rec := recs[id]
				if rec == nil {
					rec = newLogrecBuilder(id, condCount, topicCount)
					recs[id] = rec
					rec.StartFetch(tt.table.Other, tt.table.Logrec)
					defer rec.StopFetch()
				}
				rec.MatchedWith(1)
			}

			err = it.Error()
			if err != nil {
				return
			}

			it.Release()
		}
	}

	for _, rec := range recs {
		rec.MatchedWith(wildcards)
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
	rec.ok = make(chan struct{}, 1)
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
