package topicsdb

import (
	"bytes"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

func TestTopicsDb(t *testing.T) {
	logger.SetTestMode(t)

	const (
		period = 5
		count  = period * 3
	)

	topics := make([]*Topic, period)
	for i := range topics {
		t := &Topic{
			Val:  hash.FakeHash(int64(i)),
			Data: make([]byte, i+10),
		}
		_, _ = rand.Read(t.Data)
		topics[i] = t
	}

	topics4rec := func(rec int) (from, to int) {
		from = rec % (period - 3)
		to = from + 3
		return
	}

	recs := make([]*Record, count)
	for i := range recs {
		from, to := topics4rec(i)
		recs[i] = &Record{
			Id:     hash.FakeHash(int64(i)),
			BlockN: uint64(i % 5),
			Topics: topics[from:to],
		}
	}

	db := New(memorydb.New())

	t.Run("Push", func(t *testing.T) {
		assertar := assert.New(t)

		for _, rec := range recs {
			if !assertar.NoError(db.Push(rec)) {
				return
			}
		}
	})

	t.Run("Find", func(t *testing.T) {
		assertar := assert.New(t)

		for j := 0; j < period; j++ {
			from, to := topics4rec(j)
			tt := topics[from : to-1]

			conditions := make([]Condition, len(tt))
			for n, t := range tt {
				conditions[n] = NewCondition(t.Val, n)
			}

			got, err := db.Find(conditions...)
			if !assertar.NoError(err) {
				return
			}

			var expect []*Record
			for i, rec := range recs {
				if f, t := topics4rec(i); f != from || t != to {
					continue
				}

				expect = append(expect, rec)
			}

			sortById(got)
			sortById(expect)

			if !assertar.EqualValuesf(expect, got, "period=%d", j) {
				return
			}
		}
	})
}

func sortById(recs []*Record) {
	sort.Slice(recs, func(i, j int) bool {
		return bytes.Compare(
			recs[i].Id.Bytes(),
			recs[j].Id.Bytes(),
		) < 0
	})
}
