package gossip

/*
	Benchmarks for store EventHeaderData with LRU and without
*/

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/inter"
)

func TestStoreGetEventHeader(t *testing.T) {
	store := fakeLruStore()

	expect := &inter.Event{}
	expect.ClaimedTime = inter.Timestamp(rand.Int63())

	store.SetEventHeader(expect.Epoch, expect.Hash(), &expect.EventHeaderData)
	got := store.GetEventHeader(expect.Epoch, expect.Hash())

	assert.EqualValues(t, expect.EventHeaderData, *got)
}

func BenchmarkReadHeader(b *testing.B) {
	b.Run("LRU on", func(b *testing.B) {
		benchReadEventHeaderTest(b, fakeLruStore())
	})
	b.Run("LRU off", func(b *testing.B) {
		benchReadEventHeaderTest(b, fakeSimpleStore())
	})
}

func benchReadEventHeaderTest(b *testing.B, store *Store) {
	testEvent := &inter.Event{}

	store.SetEventHeader(testEvent.Epoch, testEvent.Hash(), &testEvent.EventHeaderData)

	for i := 0; i < b.N; i++ {
		_ = store.GetEventHeader(testEvent.Epoch, testEvent.Hash())
	}
}

func BenchmarkWriteHeader(b *testing.B) {
	b.Run("LRU on", func(b *testing.B) {
		benchWriteEventHeaderTest(b, fakeLruStore())
	})
	b.Run("LRU off", func(b *testing.B) {
		benchWriteEventHeaderTest(b, fakeSimpleStore())
	})
}

func benchWriteEventHeaderTest(b *testing.B, store *Store) {
	testEvent := &inter.Event{}

	for i := 0; i < b.N; i++ {
		store.SetEventHeader(testEvent.Epoch, testEvent.Hash(), &testEvent.EventHeaderData)
	}
}
