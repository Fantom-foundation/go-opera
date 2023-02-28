package topicsdb

import (
	"context"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func BenchmarkSearch(b *testing.B) {
	topics, recs, topics4rec := genTestData(1000)

	index := newTestIndex()

	for _, rec := range recs {
		err := index.Push(rec)
		require.NoError(b, err)
	}

	var query [][][]common.Hash
	for i := 0; i < len(topics); i++ {
		from, to := topics4rec(i)
		tt := topics[from : to-1]

		qq := make([][]common.Hash, len(tt))
		for pos, t := range tt {
			qq[pos] = []common.Hash{t}
		}

		query = append(query, qq)
	}

	pooled := withThreadPool{index}

	for dsc, method := range map[string]func(context.Context, idx.Block, idx.Block, [][]common.Hash) ([]*types.Log, error){
		"index":  index.FindInBlocks,
		"pooled": pooled.FindInBlocks,
	} {
		b.Run(dsc, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				qq := query[i%len(query)]
				_, err := method(nil, 0, 0xffffffff, qq)
				require.NoError(b, err)
			}
		})
	}
}
