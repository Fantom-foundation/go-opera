package topicsdb

import (
	"errors"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type (
	liteLogrecBuilder struct {
		ID          ID
		topicsCount uint8
	}
)

func newLiteLogrecBuilder(logrec ID, topicCount uint8) *liteLogrecBuilder {
	rec := &liteLogrecBuilder{
		ID:          logrec,
		topicsCount: topicCount,
	}

	return rec
}

// Fetch log record's data.
func (rec *liteLogrecBuilder) FetchLog(
	othersTable kvdb.Iteratee,
	logrecTable kvdb.Reader,
) (r *types.Log, err error) {
	r = &types.Log{
		BlockNumber: rec.ID.BlockNumber(),
		TxHash:      rec.ID.TxHash(),
		Index:       rec.ID.Index(),
		Topics:      make([]common.Hash, rec.topicsCount),
	}

	it := othersTable.NewIterator(rec.ID.Bytes(), nil)
	defer it.Release()
	for it.Next() {
		pos := extractTopicPos(it.Key())
		topic := common.BytesToHash(it.Value())
		r.Topics[pos] = topic
	}

	err = it.Error()
	if err != nil {
		return
	}

	// fields
	buf, err := logrecTable.Get(rec.ID.Bytes())
	if err != nil {
		return
	}
	offset := 0
	r.BlockHash = common.BytesToHash(buf[offset : offset+common.HashLength])
	offset += common.HashLength
	r.Data = buf[offset:]

	r.Address = common.BytesToAddress(r.Topics[0].Bytes())
	r.Topics = r.Topics[1:]

	return
}

type (
	logrecBuilder struct {
		types.Log

		ID          ID
		conditions  uint8
		topicsCount uint8

		ok          chan bool
		fetchResult error
	}
)

var ErrIsNotReady = errors.New("Is not ready yet")

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
		fetchResult: ErrIsNotReady,
	}

	return rec
}

func (rec *logrecBuilder) Build() (*types.Log, error) {
	if rec.fetchResult != nil {
		return nil, rec.fetchResult
	}

	rec.Address = common.BytesToAddress(rec.Topics[0].Bytes())
	rec.Topics = rec.Topics[1:]
	return &rec.Log, nil
}

// MatchedWith count of conditions.
func (rec *logrecBuilder) MatchedWith(count uint8) {
	if rec.conditions > count {
		rec.conditions -= count
		return
	}
	rec.conditions = 0
	if rec.ok != nil {
		rec.ok <- true
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
	defer func() {
		rec.fetchResult = err
	}()
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
