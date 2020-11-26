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

// Find wraps ForEach() for tests.
func (tt *Index) Find(topics [][]common.Hash) (all []*types.Log, err error) {
	err = tt.ForEach(topics, func(item *types.Log) (next bool) {
		all = append(all, item)
		next = true
		return
	})

	return
}

func TestIndexSearchMultyVariants(t *testing.T) {
	logger.SetTestMode(t)
	var (
		hash1 = common.BytesToHash([]byte("topic1"))
		hash2 = common.BytesToHash([]byte("topic2"))
		hash3 = common.BytesToHash([]byte("topic3"))
		hash4 = common.BytesToHash([]byte("topic4"))
		addr1 = randAddress()
		addr2 = randAddress()
		addr3 = randAddress()
		addr4 = randAddress()
	)
	testdata := []*types.Log{{
		BlockNumber: 1,
		Address:     addr1,
		Topics:      []common.Hash{hash1, hash1, hash1},
	}, {
		BlockNumber: 2,
		Address:     addr2,
		Topics:      []common.Hash{hash2, hash2, hash2},
	}, {
		BlockNumber: 998,
		Address:     addr3,
		Topics:      []common.Hash{hash3, hash3, hash3},
	}, {
		BlockNumber: 999,
		Address:     addr4,
		Topics:      []common.Hash{hash4, hash4, hash4},
	},
	}

	index := New(memorydb.New())

	for _, l := range testdata {
		err := index.Push(l)
		require.NoError(t, err)
	}

	// require.ElementsMatchf(testdata, got, "") doesn't work properly here,
	// so use check()
	check := func(require *require.Assertions, got []*types.Log) {
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

	t.Run("With no addresses", func(t *testing.T) {
		require := require.New(t)
		got, err := index.Find([][]common.Hash{
			{},
			{hash1, hash2, hash3, hash4},
			{},
			{hash1, hash2, hash3, hash4},
		})
		require.NoError(err)
		check(require, got)
	})

	t.Run("With addresses", func(t *testing.T) {
		require := require.New(t)
		got, err := index.Find([][]common.Hash{
			{addr1.Hash(), addr2.Hash(), addr3.Hash(), addr4.Hash()},
			{hash1, hash2, hash3, hash4},
			{},
			{hash1, hash2, hash3, hash4},
		})
		require.NoError(err)
		check(require, got)
	})
}

func TestIndexSearchSingleVariant(t *testing.T) {
	logger.SetTestMode(t)
	require := require.New(t)

	topics, recs, topics4rec := genTestData()

	index := New(memorydb.New())

	for _, rec := range recs {
		err := index.Push(rec)
		require.NoError(err)
	}

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

func TestIndexSearchSimple(t *testing.T) {
	logger.SetTestMode(t)
	require := require.New(t)

	var (
		hash1 = common.BytesToHash([]byte("topic1"))
		hash2 = common.BytesToHash([]byte("topic2"))
		hash3 = common.BytesToHash([]byte("topic3"))
		hash4 = common.BytesToHash([]byte("topic4"))
		addr  = randAddress()
	)
	testdata := []*types.Log{{
		BlockNumber: 1,
		Address:     addr,
		Topics:      []common.Hash{hash1},
	}, {
		BlockNumber: 2,
		Address:     addr,
		Topics:      []common.Hash{hash2},
	}, {
		BlockNumber: 998,
		Address:     addr,
		Topics:      []common.Hash{hash3},
	}, {
		BlockNumber: 999,
		Address:     addr,
		Topics:      []common.Hash{hash4},
	},
	}

	index := New(memorydb.New())

	for _, l := range testdata {
		err := index.Push(l)
		require.NoError(err)
	}

	var (
		got []*types.Log
		err error
	)

	got, err = index.Find([][]common.Hash{
		{addr.Hash()},
		{hash1},
	})
	require.NoError(err)
	require.Equal(1, len(got))

	got, err = index.Find([][]common.Hash{
		{addr.Hash()},
		{hash2},
	})
	require.NoError(err)
	require.Equal(1, len(got))

	got, err = index.Find([][]common.Hash{
		{addr.Hash()},
		{hash3},
	})
	require.NoError(err)
	require.Equal(1, len(got))
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
