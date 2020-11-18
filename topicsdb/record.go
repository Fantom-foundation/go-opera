package topicsdb

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type (
	logrecBuilder struct {
		types.Log

		ID          ID
		conditions  uint8
		topicsCount uint8

		ok    chan struct{}
		ready chan error
	}
)

func newLogrecBuilder(logrec ID, conditions uint8, topicCount uint8) *logrecBuilder {
	rec := &logrecBuilder{
		Log: types.Log{
			BlockNumber: logrec.BlockNumber(),
			TxHash:      logrec.TxHash(),
			Index:       logrec.Index(),
			Topics:      make([]common.Hash, topicCount),
		},
		ID:          logrec,
		conditions:  conditions,
		topicsCount: topicCount,
	}

	return rec
}

func (rec *logrecBuilder) Build() (r *types.Log, err error) {
	if rec.ready != nil {
		var complete bool
		err, complete = <-rec.ready
		if !complete {
			return
		}
	}

	rec.Address = common.BytesToAddress(rec.Topics[0].Bytes())
	rec.Topics = rec.Topics[1:]

	r = &rec.Log
	return
}

// MatchedWith count of conditions.
func (rec *logrecBuilder) MatchedWith(count uint8) {
	if rec.conditions > count {
		rec.conditions -= count
		return
	}
	rec.conditions = 0
	if rec.ok != nil {
		rec.ok <- struct{}{}
	}
}

// IsMatched if it is matched with the all conditions.
func (rec *logrecBuilder) IsMatched() bool {
	return rec.conditions == 0
}

// SetOtherTopic appends topic.
func (rec *logrecBuilder) SetOtherTopic(pos uint8, topic common.Hash) {
	if pos >= rec.topicsCount {
		log.Crit("inconsistent table.Others", "param", "topicN")
	}

	var empty common.Hash
	if rec.Topics[pos] != empty {
		return
	}

	rec.Topics[pos] = topic
}

// Fetch log record's data.
func (rec *logrecBuilder) Fetch(
	othersTable kvdb.Iteratee,
	logrecTable kvdb.Reader,
) (err error) {
	// others
	it := othersTable.NewIterator(rec.ID.Bytes(), nil)
	for it.Next() {
		pos := extractTopicPos(it.Key())
		topic := common.BytesToHash(it.Value())
		rec.SetOtherTopic(pos, topic)
	}

	err = it.Error()
	if err != nil {
		return
	}

	it.Release()

	// fields
	buf, err := logrecTable.Get(rec.ID.Bytes())
	if err != nil {
		return
	}
	offset := 0
	rec.BlockHash = common.BytesToHash(buf[offset : offset+common.HashLength])
	offset += common.HashLength
	rec.Data = buf[offset:]

	return
}
