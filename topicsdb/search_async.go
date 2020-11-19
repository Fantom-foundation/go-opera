package topicsdb

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (tt *Index) fetchAsync(topics [][]common.Hash, onLog func(*types.Log) (next bool)) (err error) {
	var (
		recs      = make(map[ID]*logrecBuilder)
		condCount = uint8(len(topics))
		wildcards uint8
		prefix    [prefixSize]byte
	)

	done := make(chan struct{})
	defer close(done)

	ready := make(chan *logrecBuilder, 10)

	first := true
	for pos, cond := range topics {
		if len(cond) < 1 {
			wildcards++
			continue
		}
		matched := make(map[ID]*logrecBuilder, len(recs))
		copy(prefix[common.HashLength:], posToBytes(uint8(pos)))
		for _, alternative := range cond {
			copy(prefix[:], alternative[:])
			it := tt.table.Topic.NewIterator(prefix[:], nil)
			for it.Next() {
				id := extractLogrecID(it.Key())
				topicCount := bytesToPos(it.Value())
				rec := recs[id]
				if rec == nil {
					if first {
						rec = newLogrecBuilder(id, condCount, topicCount)
						recs[id] = rec
						rec.StartFetch(tt.table.Other, tt.table.Logrec, ready, done)
					} else {
						continue
					}
				}
				rec.MatchedWith(1)
				// move rec to matched
				matched[id] = rec
				delete(recs, id)
			}

			err = it.Error()
			it.Release()
			if err != nil {
				return
			}
		}
		// clean unmatched
		for _, rec := range recs {
			rec.StopFetch()
		}
		recs = matched
		first = false
	}

	for _, rec := range recs {
		rec.MatchedWith(wildcards)
	}

	for i := 0; i < len(recs); i++ {
		rec := <-ready

		var r *types.Log
		r, err = rec.Build()
		if err != nil {
			return
		}

		if !onLog(r) {
			return
		}
	}

	return
}

// StartFetch log record's data when all conditions are ok.
func (rec *logrecBuilder) StartFetch(
	othersTable kvdb.Iteratee,
	logrecTable kvdb.Reader,
	ready chan<- *logrecBuilder,
	done <-chan struct{},
) {
	if rec.ok != nil {
		return
	}
	rec.ok = make(chan bool, 1)

	go func() {
		select {
		case ok := <-rec.ok:
			if !ok {
				return
			}
		case <-done:
			return
		}

		rec.Fetch(othersTable, logrecTable)

		select {
		case ready <- rec:
			return
		case <-done:
			return
		}
	}()
}

// StopFetch releases resources associated with StartFetch,
// so you should call StopFetch after StartFetch.
func (rec *logrecBuilder) StopFetch() {
	if rec.ok != nil {
		close(rec.ok)
		rec.ok = nil
	}
}
