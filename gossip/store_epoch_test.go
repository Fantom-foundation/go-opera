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
	store := lruStore

	expect := &inter.Event{}
	expect.ClaimedTime = inter.Timestamp(rand.Int63())

	store.SetEventHeader(expect.Epoch, expect.Hash(), &expect.EventHeaderData)
	got := store.GetEventHeader(expect.Epoch, expect.Hash())

	assert.EqualValues(t, expect.EventHeaderData, *got)
}

func BenchmarkReadHeader(b *testing.B) {
	testStore = lruStore
	b.Run("LRUon", benchReadEventHeaderTest)

	testStore = simpleStore
	b.Run("LRUoff", benchReadEventHeaderTest)
}

func benchReadEventHeaderTest(b *testing.B) {
	testEvent := &inter.Event{}

	testStore.SetEventHeader(testEvent.Epoch, testEvent.Hash(), &testEvent.EventHeaderData)

	for i := 0; i < b.N; i++ {
		_ = testStore.GetEventHeader(testEvent.Epoch, testEvent.Hash())
	}
}

func BenchmarkWriteHeader(b *testing.B) {
	testStore = lruStore
	b.Run("LRUon", benchWriteEventHeaderTest)

	testStore = simpleStore
	b.Run("LRUoff", benchWriteEventHeaderTest)
}

func benchWriteEventHeaderTest(b *testing.B) {
	testEvent := &inter.Event{}

	for i := 0; i < b.N; i++ {
		testStore.SetEventHeader(testEvent.Epoch, testEvent.Hash(), &testEvent.EventHeaderData)
	}
}
