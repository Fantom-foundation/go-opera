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

			conditions := make([]Condition, len(tt))
			for pos, t := range tt {
				conditions[pos] = NewCondition(t.Topic, uint8(pos))
			}

			got, err := db.Find(conditions...)
			if !assertar.NoError(err) {
				return
			}

			var expect []*Logrec
			for j, rec := range recs {
				if f, t := topics4rec(j); f != from || t != to {
					continue
				}

				expect = append(expect, rec)
			}

			sortById(got)
			sortById(expect)

			if !assertar.EqualValuesf(expect, got, "step %d", i) {
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
	topics []*Topic,
	recs []*Logrec,
	topics4rec func(rec int) (from, to int),
) {
	const (
		period = 5
		count  = period * 3
	)

	topics = make([]*Topic, period)
	for i := range topics {
		t := &Topic{
			Topic: hash.FakeHash(int64(i)),
			Data:  make([]byte, i+10),
		}
		_, _ = rand.Read(t.Data)
		topics[i] = t
	}

	topics4rec = func(rec int) (from, to int) {
		from = rec % (period - 3)
		to = from + 3
		return
	}

	recs = make([]*Logrec, count)
	for i := range recs {
		from, to := topics4rec(i)
		recs[i] = &Logrec{
			Id:     hash.FakeHash(int64(i)),
			BlockN: uint64(i % 5),
			Topics: topics[from:to],
		}
	}

	return
}

func sortById(recs []*Logrec) {
	sort.Slice(recs, func(i, j int) bool {
		return bytes.Compare(
			recs[i].Id.Bytes(),
			recs[j].Id.Bytes(),
		) < 0
	})
}
