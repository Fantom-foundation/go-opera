package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
)

const conditionSize = lenHash + lenInt32

// Condition to search log records by.
type Condition [conditionSize]byte

// NewCondition from topic and its position in log record.
func NewCondition(topic common.Hash, position int) Condition {
	var c Condition

	copy(c[:], topic.Bytes())
	copy(c[lenHash:], bigendian.Int32ToBytes(uint32(position)))

	return c
}

func (tt *TopicsDb) fetchSync(cc ...Condition) (res []*Logrec, err error) {
	recs := make(map[common.Hash]*logrecBuilder)

	for _, cond := range cc {
		it := tt.table.Topic.NewIteratorWithPrefix(cond[:])
		for it.Next() {
			key := it.Key()
			id := extractLogrecId(key)
			blockN := extractBlockN(key)
			topicCount := bigendian.BytesToInt32(it.Value())
			rec := recs[id]
			if rec == nil {
				rec = newLogrecBuilder(len(cc), id, blockN, topicCount)
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

		it := tt.table.Logrec.NewIteratorWithPrefix(rec.id.Bytes())
		for it.Next() {
			n := extractTopicN(it.Key())
			rec.SetTopic(n, it.Value())
		}

		err = it.Error()
		if err != nil {
			return
		}

		it.Release()

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

// Position getter.
func (cond Condition) Position() uint32 {
	return bigendian.BytesToInt32(cond[lenHash:])
}

// Topic getter.
func (cond Condition) Topic() common.Hash {
	return common.BytesToHash(cond[:lenHash])
}
