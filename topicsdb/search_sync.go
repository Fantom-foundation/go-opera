package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const prefixSize = hashSize + uint8Size

func (tt *Index) fetchSync(topics [][]common.Hash, onLog func(*types.Log) (next bool)) (err error) {
	var (
		recs      = make(map[ID]*logrecBuilder)
		condCount = uint8(len(topics))
		wildcards uint8
		prefix    [prefixSize]byte
	)
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
					} else {
						continue
					}
				}
				rec.MatchedWith(1)
				matched[id] = rec
			}

			first = false
			err = it.Error()
			it.Release()
			if err != nil {
				return
			}
		}
		recs = matched
	}

	for _, rec := range recs {
		rec.MatchedWith(wildcards)
		if !rec.IsMatched() {
			continue
		}

		err = rec.Fetch(tt.table.Other, tt.table.Logrec)
		if err != nil {
			return
		}

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
