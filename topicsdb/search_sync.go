package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const prefixSize = hashSize + uint8Size

func (tt *Index) fetchSync(topics [][]common.Hash) (res []*types.Log, err error) {
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

		err = rec.Fetch(tt.table.Other, tt.table.Logrec)
		if err != nil {
			return
		}

		var r *types.Log
		r, err = rec.Build()
		if err != nil {
			return
		}
		res = append(res, r)
	}

	return
}
