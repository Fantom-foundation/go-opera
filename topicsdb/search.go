package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
)

const conditionSize = lenHash + lenInt32

type Condition [conditionSize]byte

func NewCondition(topic common.Hash, n int) Condition {
	var c Condition

	copy(c[:], topic.Bytes())
	copy(c[lenHash:], bigendian.Int32ToBytes(uint32(n)))

	return c
}

func (tt *TopicsDb) fetchSync(cc ...Condition) (res []*Record, err error) {
	recs := make(map[common.Hash]*recordBuilder)

	for _, cond := range cc {
		it := tt.table.Topic.NewIteratorWithPrefix(cond[:])
		for it.Next() {
			key := it.Key()
			id := extractRecId(key)
			blockN := extractBlockN(key)
			topicCount := bigendian.BytesToInt32(it.Value())
			rec := recs[id]
			if rec == nil {
				rec = newRecordBuilder(len(cc), id, blockN, topicCount)
				recs[id] = rec
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
		if !rec.AllConditionsOK() {
			continue
		}

		it := tt.table.Record.NewIteratorWithPrefix(rec.id.Bytes())
		for it.Next() {
			n := extractTopicN(it.Key())
			rec.SetTopic(n, it.Value())
		}

		err = it.Error()
		if err != nil {
			return
		}

		it.Release()

		var r *Record
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

func (cond Condition) N() uint32 {
	return bigendian.BytesToInt32(cond[lenHash:])
}

func (cond Condition) Val() common.Hash {
	return common.BytesToHash(cond[:lenHash])
}
