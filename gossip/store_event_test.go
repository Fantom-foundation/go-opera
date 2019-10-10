package gossip

/*
	Benchmarks for store Events with LRU and without
*/

import (
	"github.com/Fantom-foundation/go-lachesis/inter"
	"testing"
)

func BenchmarkReadEventWithLRU(b *testing.B) {
	StoreConfig = &ExtendedStoreConfig{
		EventsCacheSize:        100,
		EventsHeadersCacheSize: 10000,
	}

	benchReadEventTest(b)
}

func BenchmarkReadEventWithoutLRU(b *testing.B) {
	StoreConfig = nil

	benchReadEventTest(b)
}

func benchReadEventTest(b *testing.B) {
	testEvent := &inter.Event{}
	store := NewMemStore()

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
	StoreConfig = &ExtendedStoreConfig{
		EventsCacheSize:        100,
		EventsHeadersCacheSize: 10000,
	}

	benchWriteEventTest(b)
}

func BenchmarkWriteEventWithoutLRU(b *testing.B) {
	StoreConfig = nil

	benchWriteEventTest(b)
}

func benchWriteEventTest(b *testing.B) {
	testEvent := &inter.Event{}
	store := NewMemStore()

	for i := 0; i < b.N; i++ {
		store.SetEvent(testEvent)
	}
}

func BenchmarkHasEventExistsWithLRU(b *testing.B) {
	StoreConfig = &ExtendedStoreConfig{
		EventsCacheSize:        100,
		EventsHeadersCacheSize: 10000,
	}

	benchHasEventExistsTest(b)
}

func BenchmarkHasEventExistsWithoutLRU(b *testing.B) {
	StoreConfig = nil

	benchHasEventExistsTest(b)
}

func BenchmarkHasEventAbsentWithLRU(b *testing.B) {
	StoreConfig = &ExtendedStoreConfig{
		EventsCacheSize:        100,
		EventsHeadersCacheSize: 10000,
	}

	benchHasEventAbsentTest(b)
}

func BenchmarkHasEventAbsentWithoutLRU(b *testing.B) {
	StoreConfig = nil

	benchHasEventAbsentTest(b)
}

func benchHasEventExistsTest(b *testing.B) {
	testEvent := &inter.Event{}
	store := NewMemStore()

	store.SetEvent(testEvent)

	hev := testEvent.Hash()
	for i := 0; i < b.N; i++ {
		if !store.HasEvent(hev) {
			b.Fatalf("Not exists saved event\n")
		}
	}
}

func benchHasEventAbsentTest(b *testing.B) {
	testEvent := &inter.Event{}
	store := NewMemStore()

	store.DeleteEvent(testEvent.Epoch, testEvent.Hash())

	hev := testEvent.Hash()
	for i := 0; i < b.N; i++ {
		if store.HasEvent(hev) {
			b.Fatalf("Exists absent event\n")
		}
	}
}
