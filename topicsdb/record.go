package topicsdb

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type (
	logrec struct {
		ID          ID
		topicsCount uint8
	}
)

func newLogrec(rec ID, topicCount uint8) *logrec {
	return &logrec{
		ID:          rec,
		topicsCount: topicCount,
	}
}

// FetchLog record's data.
func (rec *logrec) FetchLog(
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
