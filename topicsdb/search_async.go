package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
)

func (tt *TopicsDb) fetchAsync(cc ...Condition) (res []*Logrec, err error) {
	recs := make(map[common.Hash]*logrecBuilder)

	for _, cond := range cc {
		it := tt.table.Topic.NewIteratorWithPrefix(cond[:])
		for it.Next() {
			key := it.Key()
			id := extractRecId(key)
			blockN := extractBlockN(key)
			topicCount := bigendian.BytesToInt32(it.Value())
			rec := recs[id]
			if rec == nil {
				rec = newLogrecBuilder(len(cc), id, blockN, topicCount)
				recs[id] = rec
				rec.StartFetch(tt.table.Record.NewIteratorWithPrefix)
				defer rec.StopFetch()
			} else {
				rec.SetParams(blockN, topicCount)
			}
			rec.ConditionOK(cond)
		}

		err = it.Error()
		if err != nil {
			return
		}

		it.Release()
	}

	for _, rec := range recs {
		var r *Logrec
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

// StartFetch record's data when all conditions are ok.
func (rec *logrecBuilder) StartFetch(fetch func(prefix []byte) ethdb.Iterator) {
	if rec.ok != nil {
		return
	}
	rec.ok = make(chan struct{})
	rec.ready = make(chan error)

	go func() {
		defer close(rec.ready)

		_, conditions := <-rec.ok
		if !conditions {
			return
		}

		it := fetch(rec.id.Bytes())
		defer it.Release()

		for it.Next() {
			n := extractTopicN(it.Key())
			rec.SetTopic(n, it.Value())
		}

		rec.ready <- it.Error()
	}()
}

// StopFetch releases resources associated with StartFetch,
// so code should call StopFetch after StartFetch.
func (rec *logrecBuilder) StopFetch() {
	if rec.ok != nil {
		close(rec.ok)
		rec.ok = nil
	}
	rec.ready = nil
}
