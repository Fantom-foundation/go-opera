package topicsdb

import (
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/logger"
)

func TestIndexSearchMultyVariants(t *testing.T) {
	logger.SetTestMode(t)
	var (
		hash1 = common.BytesToHash([]byte("topic1"))
		hash2 = common.BytesToHash([]byte("topic2"))
		hash3 = common.BytesToHash([]byte("topic3"))
		hash4 = common.BytesToHash([]byte("topic4"))
	)
	testdata := []*types.Log{{
		BlockNumber: 1,
		Address:     randAddress(),
		Topics:      []common.Hash{hash1},
	}, {
		BlockNumber: 2,
		Address:     randAddress(),
		Topics:      []common.Hash{hash2},
	}, {
		BlockNumber: 998,
		Address:     randAddress(),
		Topics:      []common.Hash{hash3},
	}, {
		BlockNumber: 999,
		Address:     randAddress(),
		Topics:      []common.Hash{hash4},
	},
	}

	index := New(memorydb.New())

	for _, l := range testdata {
		err := index.Push(l)
		require.NoError(t, err)
	}

	find := func(t *testing.T) {
		require := require.New(t)
		got, err := index.Find([][]common.Hash{{}, {hash1, hash2, hash3, hash4}})
		require.NoError(err)
		// require.ElementsMatchf(testdata, got, "") doesn't work properly here, so:
		count := 0
		for _, a := range got {
			for _, b := range testdata {
				if b.Address == a.Address {
					require.ElementsMatch(a.Topics, b.Topics)
					count++
					break
				}
			}
		}
		require.Equal(len(testdata), count)
	}

	t.Run("Find lazy", func(t *testing.T) {
		index.fetchMethod = index.fetchLazy
		find(t)
	})
}

func TestIndexSearchSingleVariant(t *testing.T) {
	logger.SetTestMode(t)

	topics, recs, topics4rec := genTestData()

	index := New(memorydb.New())

	for _, rec := range recs {
		err := index.Push(rec)
		require.NoError(t, err)
	}

	find := func(t *testing.T) {
		require := require.New(t)

		for i := 0; i < len(topics); i++ {
			from, to := topics4rec(i)
			tt := topics[from : to-1]

			qq := make([][]common.Hash, len(tt)+1)
			for pos, t := range tt {
				qq[pos+1] = []common.Hash{t}
			}

			got, err := index.Find(qq)
			require.NoError(err)

			var expect []*types.Log
			for j, rec := range recs {
				if f, t := topics4rec(j); f != from || t != to {
					continue
				}
				expect = append(expect, rec)
			}

			require.ElementsMatchf(expect, got, "step %d", i)
		}
	}

	t.Run("Find lazy", func(t *testing.T) {
		index.fetchMethod = index.fetchLazy
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
			Address:     randAddress(),
			Topics:      topics[from:to],
			Data:        make([]byte, i),
		}
		_, _ = rand.Read(r.Data)
		recs[i] = r
	}

	return
}

func randAddress() (addr common.Address) {
	n, err := rand.Read(addr[:])
	if err != nil {
		panic(err)
	}
	if n != common.AddressLength {
		panic("address is not filled")
	}
	return
}
