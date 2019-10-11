package gossip

/*
	Benchmarks for store EventHeaderData with LRU and without
*/

import (
	"math/rand"
	"testing"

	lru "github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-lachesis/inter"
)

func TestStoreGetEventHeader(t *testing.T) {
	store := NewMemStore()
	store.cache.Events, _ = lru.New(100)
	store.cache.EventsHeaders, _ = lru.New(100)

	expect := &inter.Event{}
	expect.ClaimedTime = inter.Timestamp(rand.Int63())

	store.SetEventHeader(expect.Epoch, expect.Hash(), &expect.EventHeaderData)
	got := store.GetEventHeader(expect.Epoch, expect.Hash())

	if got.Hash() != expect.EventHeaderData.Hash() {
		t.Error("Error save/restore EventHeader with LRU" )
	}
}

func BenchmarkReadHeaderWithLRU(b *testing.B) {
	store := NewMemStore()
	store.cache.Events, _ = lru.New(100)
	store.cache.EventsHeaders, _ = lru.New(100)

	benchReadEventHeaderTest(b, store)
}

func BenchmarkReadHeaderWithoutLRU(b *testing.B) {
	store := NewMemStore()
	store.cache.Events = nil
	store.cache.EventsHeaders = nil

	benchReadEventHeaderTest(b, store)
}

func benchReadEventHeaderTest(b *testing.B, store *Store) {
	testEvent := &inter.Event{}

	store.SetEventHeader(testEvent.Epoch, testEvent.Hash(), &testEvent.EventHeaderData)

	key := testEvent.EventHeaderData.Hash().Bytes()
	for i := 0; i < b.N; i++ {
		hev := store.GetEventHeader(testEvent.Epoch, testEvent.Hash())
		if string(hev.Hash().Bytes()) != string(key) {
			b.Fatalf("Stored event header '%s' not equal original '%s'\n", string(hev.Hash().Bytes()), string(key))
		}
	}
}

func BenchmarkWriteHeaderWithLRU(b *testing.B) {
	store := NewMemStore()
	store.cache.Events, _ = lru.New(100)
	store.cache.EventsHeaders, _ = lru.New(100)

	benchWriteEventHeaderTest(b, store)
}

func BenchmarkWriteHeaderWithoutLRU(b *testing.B) {
	store := NewMemStore()
	store.cache.Events = nil
	store.cache.EventsHeaders = nil

	benchWriteEventHeaderTest(b, store)
}

func benchWriteEventHeaderTest(b *testing.B, store *Store) {
	testEvent := &inter.Event{}

	for i := 0; i < b.N; i++ {
		store.SetEventHeader(testEvent.Epoch, testEvent.Hash(), &testEvent.EventHeaderData)
	}
}
