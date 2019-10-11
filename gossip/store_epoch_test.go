package gossip

/*
	Benchmarks for store EventHeaderData with LRU and without
*/

import (
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/inter"
)

func TestStoreGetEventHeader(t *testing.T) {
	store := lruStore

	expect := &inter.Event{}
	expect.ClaimedTime = inter.Timestamp(rand.Int63())

	store.SetEventHeader(expect.Epoch, expect.Hash(), &expect.EventHeaderData)
	got := store.GetEventHeader(expect.Epoch, expect.Hash())

	if got.Hash() != expect.EventHeaderData.Hash() {
		t.Error("Error save/restore EventHeader with LRU" )
	}
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

	key := testEvent.EventHeaderData.Hash().Bytes()
	for i := 0; i < b.N; i++ {
		hev := testStore.GetEventHeader(testEvent.Epoch, testEvent.Hash())
		if string(hev.Hash().Bytes()) != string(key) {
			b.Fatalf("Stored event header '%s' not equal original '%s'\n", string(hev.Hash().Bytes()), string(key))
		}
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
