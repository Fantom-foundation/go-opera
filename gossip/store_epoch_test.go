package gossip

/*
	Benchmarks for store EventHeaderData with LRU and without
*/

import (
	"github.com/Fantom-foundation/go-lachesis/inter"
	"testing"
)

func BenchmarkReadHeaderWithLRU(b *testing.B) {
	StoreConfig = &ExtendedStoreConfig{
		EventsCacheSize:        100,
		EventsHeadersCacheSize: 10000,
	}

	benchReadEventHeaderTest(b)
}

func BenchmarkReadHeaderWithoutLRU(b *testing.B) {
	StoreConfig = nil

	benchReadEventHeaderTest(b)
}

func benchReadEventHeaderTest(b *testing.B) {
	testEvent := &inter.Event{}
	store := NewMemStore()

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
	StoreConfig = &ExtendedStoreConfig{
		EventsCacheSize:        100,
		EventsHeadersCacheSize: 10000,
	}

	benchWriteEventHeaderTest(b)
}

func BenchmarkWriteHeaderWithoutLRU(b *testing.B) {
	StoreConfig = nil

	benchWriteEventHeaderTest(b)
}

func benchWriteEventHeaderTest(b *testing.B) {
	testEvent := &inter.Event{}
	store := NewMemStore()

	for i := 0; i < b.N; i++ {
		store.SetEventHeader(testEvent.Epoch, testEvent.Hash(), &testEvent.EventHeaderData)
	}
}
