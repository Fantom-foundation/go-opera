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
	othersTable kvdb.Reader,
	logrecTable kvdb.Reader,
) (r *types.Log, err error) {
	r = &types.Log{
		BlockNumber: rec.ID.BlockNumber(),
		TxHash:      rec.ID.TxHash(),
		Index:       rec.ID.Index(),
		Topics:      make([]common.Hash, rec.topicsCount),
	}

	var buf []byte

	// topics
	buf, err = othersTable.Get(rec.ID.Bytes())
	if err != nil {
		return
	}
	pos := 0
	for offset := 0; offset < len(buf); offset += common.HashLength {
		topic := common.BytesToHash(buf[offset : offset+common.HashLength])
		r.Topics[pos] = topic
		pos++
	}

	// fields
	buf, err = logrecTable.Get(rec.ID.Bytes())
	if err != nil {
		return
	}
	offset := 0
	r.BlockHash = common.BytesToHash(buf[offset : offset+common.HashLength])
	offset += common.HashLength
	r.Address = common.BytesToAddress(buf[offset : offset+common.AddressLength])
	offset += common.AddressLength
	r.Data = buf[offset:]

	return
}
