package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/inter"
	"testing"
)

func BenchmarkReadWithLRU(b *testing.B) {
	StoreConfig = &ExtendedStoreConfig{
		EventsCacheSize:        100,
		EventsHeadersCacheSize: 10000,
	}

	benchReadTest(b)
}

func BenchmarkReadWithoutLRU(b *testing.B) {
	StoreConfig = nil

	benchReadTest(b)
}

func benchReadTest(b *testing.B) {
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

func BenchmarkWriteWithLRU(b *testing.B) {
	StoreConfig = &ExtendedStoreConfig{
		EventsCacheSize:        100,
		EventsHeadersCacheSize: 10000,
	}

	benchWriteTest(b)
}

func BenchmarkWriteWithoutLRU(b *testing.B) {
	StoreConfig = nil

	benchWriteTest(b)
}

func benchWriteTest(b *testing.B) {
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
