package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
)

const conditionSize = lenHash + lenUint8

// Condition to search log records by.
type Condition [conditionSize]byte

// NewCondition from topic and its position in log record.
func NewCondition(topic common.Hash, position uint8) Condition {
	var c Condition

	copy(c[:], topic.Bytes())
	copy(c[lenHash:], posToBytes(position))

	return c
}

func (tt *TopicsDb) fetchSync(cc ...Condition) (res []*Logrec, err error) {
	if len(cc) > MaxCount {
		err = ErrTooManyTopics
		return
	}

	recs := make(map[common.Hash]*logrecBuilder)

	conditions := uint8(len(cc))
	for _, cond := range cc {
		it := tt.table.Topic.NewIteratorWithPrefix(cond[:])
		for it.Next() {
			key := it.Key()
			id := extractLogrecID(key)
			blockN := extractBlockN(key)
			topicCount := bytesToPos(it.Value())
			rec := recs[id]
			if rec == nil {
				rec = newLogrecBuilder(conditions, id, blockN, topicCount)
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
			pos := extractTopicPos(it.Key())
			rec.SetTopic(pos, it.Value())
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
