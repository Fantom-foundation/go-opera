package topicsdb

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
)

func BenchmarkSearch(b *testing.B) {
	topics, recs, topics4rec := genTestData()

	mem := memorydb.New()
	mem.SetDelay(1 * time.Millisecond)
	db := New(mem)

	for _, rec := range recs {
		if err := db.Push(rec); err != nil {
			b.Fatal(err)
		}
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

	b.Run("Sync", func(b *testing.B) {
		db.fetchMethod = db.fetchSync
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			qq := query[i%len(query)]
			_, err := db.Find(qq)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Async", func(b *testing.B) {
		db.fetchMethod = db.fetchAsync
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			qq := query[i%len(query)]
			_, err := db.Find(qq)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

}
