package gossip

/*
	Benchmarks for store Events with LRU and without
*/

import (
	"github.com/Fantom-foundation/go-lachesis/inter"
	lru "github.com/hashicorp/golang-lru"
	"testing"
)

func BenchmarkReadEventWithLRU(b *testing.B) {
	store := NewMemStore()
	store.cache.Events, _ = lru.New(100)
	store.cache.EventsHeaders, _ = lru.New(100)

	benchReadEventTest(b, store)
}

func BenchmarkReadEventWithoutLRU(b *testing.B) {
	store := NewMemStore()
	store.cache.Events = nil
	store.cache.EventsHeaders = nil

	benchReadEventTest(b, store)
}

func benchReadEventTest(b *testing.B, store *Store) {
	testEvent := &inter.Event{}

	store.SetEvent(testEvent)

	key := testEvent.Hash().Bytes()
	for i := 0; i < b.N; i++ {
		ev := store.GetEvent(testEvent.Hash())
		if string(ev.Hash().Bytes()) != string(key) {
			b.Fatalf("Stored event '%s' not equal original '%s'\n", string(ev.Hash().Bytes()), string(key))
		}
	}
}

func BenchmarkWriteEventWithLRU(b *testing.B) {
	store := NewMemStore()
	store.cache.Events, _ = lru.New(100)
	store.cache.EventsHeaders, _ = lru.New(100)

	benchWriteEventTest(b, store)
}

func BenchmarkWriteEventWithoutLRU(b *testing.B) {
	store := NewMemStore()
	store.cache.Events = nil
	store.cache.EventsHeaders = nil

	benchWriteEventTest(b, store)
}

func benchWriteEventTest(b *testing.B, store *Store) {
	testEvent := &inter.Event{}

	for i := 0; i < b.N; i++ {
		store.SetEvent(testEvent)
	}
}

func BenchmarkHasEventExistsWithLRU(b *testing.B) {
	store := NewMemStore()
	store.cache.Events, _ = lru.New(100)
	store.cache.EventsHeaders, _ = lru.New(100)

	benchHasEventExistsTest(b, store)
}

func BenchmarkHasEventExistsWithoutLRU(b *testing.B) {
	store := NewMemStore()
	store.cache.Events = nil
	store.cache.EventsHeaders = nil

	benchHasEventExistsTest(b, store)
}

func BenchmarkHasEventAbsentWithLRU(b *testing.B) {
	store := NewMemStore()
	store.cache.Events, _ = lru.New(100)
	store.cache.EventsHeaders, _ = lru.New(100)

	benchHasEventAbsentTest(b, store)
}

func BenchmarkHasEventAbsentWithoutLRU(b *testing.B) {
	store := NewMemStore()
	store.cache.Events = nil
	store.cache.EventsHeaders = nil

	benchHasEventAbsentTest(b, store)
}

func benchHasEventExistsTest(b *testing.B, store *Store) {
	testEvent := &inter.Event{}

	store.SetEvent(testEvent)

	hev := testEvent.Hash()
	for i := 0; i < b.N; i++ {
		if !store.HasEvent(hev) {
			b.Fatalf("Not exists saved event\n")
		}
	}
}

func benchHasEventAbsentTest(b *testing.B, store *Store) {
	testEvent := &inter.Event{}

	store.DeleteEvent(testEvent.Epoch, testEvent.Hash())

	hev := testEvent.Hash()
	for i := 0; i < b.N; i++ {
		if store.HasEvent(hev) {
			b.Fatalf("Exists absent event\n")
		}
	}
}
