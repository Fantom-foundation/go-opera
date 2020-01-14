package topicsdb

import (
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

func TestTopicsDb(t *testing.T) {
	logger.SetTestMode(t)

	topics, recs, topics4rec := genTestData()

	db := New(memorydb.New())

	t.Run("Push", func(t *testing.T) {
		assertar := assert.New(t)

		for _, rec := range recs {
			if !assertar.NoError(db.Push(rec)) {
				return
			}
		}
	})

	find := func(t *testing.T) {
		assertar := assert.New(t)

		for i := 0; i < len(topics); i++ {
			from, to := topics4rec(i)
			tt := topics[from : to-1]

			qq := make([][]common.Hash, len(tt))
			for pos, t := range tt {
				qq[pos] = []common.Hash{t}
			}

			got, err := db.Find(qq)
			if !assertar.NoError(err) {
				return
			}

			var expect []*types.Log
			for j, rec := range recs {
				if f, t := topics4rec(j); f != from || t != to {
					continue
				}

				expect = append(expect, rec)
			}

			if !assertar.ElementsMatchf(expect, got, "step %d", i) {
				return
			}
		}
	}

	t.Run("Find sync", func(t *testing.T) {
		db.fetchMethod = db.fetchSync
		find(t)
	})

	t.Run("Find async", func(t *testing.T) {
		db.fetchMethod = db.fetchAsync
		find(t)
	})

}

func genTestData() (
	topics []common.Hash,
	recs []*types.Log,
	topics4rec func(rec int) (from, to int),
) {
	const (
		period = 5
		count  = period * 3
	)

	topics = make([]common.Hash, period)
	for i := range topics {
		topics[i] = hash.FakeHash(int64(i))
	}

	topics4rec = func(rec int) (from, to int) {
		from = rec % (period - 3)
		to = from + 3
		return
	}

	recs = make([]*types.Log, count)
	for i := range recs {
		from, to := topics4rec(i)
		r := &types.Log{
			BlockNumber: uint64(i / period),
			BlockHash:   hash.FakeHash(int64(i / period)),
			TxHash:      hash.FakeHash(int64(i % period)),
			Index:       uint(i % period),
			Address:     common.Address{0x1, 0x2, 0xff, 0x0},
			Topics:      topics[from:to],
			Data:        make([]byte, i),
		}
		_, _ = rand.Read(r.Data)
		recs[i] = r
	}

	return
}
