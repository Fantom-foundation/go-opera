package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
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

	r = &rec.Log
	return
}

// MatchedWith condition.
func (rec *logrecBuilder) MatchedWith(cond Condition) {
	rec.conditions--
	if rec.conditions == 0 && rec.ok != nil {
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
	othersTable ethdb.Iteratee,
	logrecTable ethdb.KeyValueReader,
) (err error) {
	// others
	it := othersTable.NewIteratorWithPrefix(rec.ID.Bytes())
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
	rec.Data, err = logrecTable.Get(rec.ID.Bytes())
	if err != nil {
		return
	}

	return
}
