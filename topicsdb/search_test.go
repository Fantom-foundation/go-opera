package topicsdb

import (
	"testing"
	"time"

	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
)

func BenchmarkSearch(b *testing.B) {
	topics, recs, topics4rec := genTestData()

	mem := memorydb.New()
	mem.SetDelay(1 * time.Millisecond)
	db := New(mem)

	for _, rec := range recs {
		err := db.Push(rec)
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

	b.Run("Lazy", func(b *testing.B) {
		// db.fetchMethod = db.fetchLazy
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			qq := query[i%len(query)]
			_, err := db.Find(qq)
			require.NoError(b, err)
		}
	})
}
