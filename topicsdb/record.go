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
	logrecTable kvdb.Reader,
) (r *types.Log, err error) {
	r = &types.Log{
		BlockNumber: rec.ID.BlockNumber(),
		TxHash:      rec.ID.TxHash(),
		Index:       rec.ID.Index(),
		Topics:      make([]common.Hash, rec.topicsCount),
	}

	var (
		buf    []byte
		offset int
	)
	buf, err = logrecTable.Get(rec.ID.Bytes())
	if err != nil {
		return
	}

	// topics
	for i := 0; i < len(r.Topics); i++ {
		r.Topics[i] = common.BytesToHash(buf[offset : offset+common.HashLength])
		offset += common.HashLength
	}

	// fields
	r.BlockHash = common.BytesToHash(buf[offset : offset+common.HashLength])
	offset += common.HashLength
	r.Address = common.BytesToAddress(buf[offset : offset+common.AddressLength])
	offset += common.AddressLength
	r.Data = buf[offset:]

	return
}
