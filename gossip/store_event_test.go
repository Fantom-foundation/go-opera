package gossip

/*
	Benchmarks for store Events with LRU and without
*/

import (
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
)

func fakeLruStore() *Store {
	return NewMemStore()
}

func fakeSimpleStore() *Store {
	store := NewMemStore()
	store.cache.Events = nil
	store.cache.EventsHeaders = nil
	store.cache.Blocks = nil
	store.cache.PackInfos = nil
	store.cache.TxPositions = nil
	store.cache.Receipts = nil

	return store
}

func TestCorrectCacheWorkForEvent(t *testing.T) {
	store := fakeLruStore()

	expect := &inter.Event{}
	expect.ClaimedTime = inter.Timestamp(rand.Int63())

	store.SetEvent(expect)
	got := store.GetEvent(expect.Hash())

	assert.EqualValues(t, expect, got)
}

func BenchmarkReadEvent(b *testing.B) {
	b.Run("LRU on", func(b *testing.B) {
		benchReadEventTest(b, fakeLruStore())
	})
	b.Run("LRU off", func(b *testing.B) {
		benchReadEventTest(b, fakeSimpleStore())
	})
}

func benchReadEventTest(b *testing.B, store *Store) {
	expect := createFakeEvent()
	if store.cache.Events != nil {
		store.cache.Events.Purge()
	}

	store.SetEvent(expect)

	for i := 0; i < b.N; i++ {
		_ = store.GetEvent(expect.Hash())
	}
}

func BenchmarkWriteEvent(b *testing.B) {
	b.Run("LRU on", func(b *testing.B) {
		benchWriteEventTest(b, fakeLruStore())
	})
	b.Run("LRU off", func(b *testing.B) {
		benchWriteEventTest(b, fakeSimpleStore())
	})
}

func benchWriteEventTest(b *testing.B, store *Store) {
	expect := &inter.Event{}

	for i := 0; i < b.N; i++ {
		store.SetEvent(expect)
	}
}

func BenchmarkHasEvent(b *testing.B) {
	b.Run("LRU on Exists", func(b *testing.B) {
		benchHasEventExistsTest(b, fakeLruStore())
	})
	b.Run("LRU off Exists", func(b *testing.B) {
		benchHasEventExistsTest(b, fakeSimpleStore())
	})
	b.Run("LRU on Absent", func(b *testing.B) {
		benchHasEventAbsentTest(b, fakeLruStore())
	})
	b.Run("LRU off Absent", func(b *testing.B) {
		benchHasEventAbsentTest(b, fakeSimpleStore())
	})
}

func benchHasEventExistsTest(b *testing.B, store *Store) {
	expect := &inter.Event{}

	store.SetEvent(expect)

	hev := expect.Hash()
	for i := 0; i < b.N; i++ {
		_ = store.HasEvent(hev)
	}
}

func benchHasEventAbsentTest(b *testing.B, store *Store) {
	expect := &inter.Event{}

	store.DeleteEvent(expect.Epoch, expect.Hash())

	hev := expect.Hash()
	for i := 0; i < b.N; i++ {
		_ = store.HasEvent(hev)
	}
}

func createFakeEvent() *inter.Event {
	d := &inter.Event{
		EventHeader: inter.EventHeader{
			EventHeaderData: inter.EventHeaderData{
				Parents: hash.Events{},
				Extra:   make([]byte, 0),
			},
			Sig: make([]byte, 0),
		},
		Transactions: types.Transactions{},
	}

	return d
}
