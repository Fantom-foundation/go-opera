package gossip

/*
	Benchmarks for store Events with LRU and without
*/

import (
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/inter"
)

var (
	lruStore *Store
	simpleStore *Store

	testStore *Store
)

func init() {
	lruStore = NewMemStore()
	simpleStore = NewMemStore()
	simpleStore.cache.Events = nil
	simpleStore.cache.EventsHeaders = nil
	simpleStore.cache.Blocks = nil
	simpleStore.cache.PackInfos = nil
	simpleStore.cache.TxPositions = nil
	simpleStore.cache.Receipts = nil
}

func TestCorrectCacheWorkForEvent(t *testing.T) {
	store := lruStore

	testEvent := &inter.Event{}
	testEvent.ClaimedTime = inter.Timestamp(rand.Int63())

	store.SetEvent(testEvent)
	e := store.GetEvent(testEvent.Hash())

	if e.Hash() != testEvent.Hash() {
		t.Error("Error save/restore Event with LRU" )
	}
}

func BenchmarkReadEvent(b *testing.B) {
	testStore = lruStore
	b.Run("LRUon", benchReadEventTest)

	testStore = simpleStore
	b.Run("LRUoff", benchReadEventTest)
}

func benchReadEventTest(b *testing.B) {
	testEvent := &inter.Event{}

	testStore.SetEvent(testEvent)

	key := testEvent.Hash().Bytes()
	for i := 0; i < b.N; i++ {
		ev := testStore.GetEvent(testEvent.Hash())
		if string(ev.Hash().Bytes()) != string(key) {
			b.Fatalf("Stored event '%s' not equal original '%s'\n", string(ev.Hash().Bytes()), string(key))
		}
	}
}

func BenchmarkWriteEvent(b *testing.B) {
	testStore = lruStore
	b.Run("LRUon", benchWriteEventTest)

	testStore = simpleStore
	b.Run("LRUoff", benchWriteEventTest)
}

func benchWriteEventTest(b *testing.B) {
	testEvent := &inter.Event{}

	for i := 0; i < b.N; i++ {
		testStore.SetEvent(testEvent)
	}
}

func BenchmarkHasEvent(b *testing.B) {
	testStore = lruStore
	b.Run("LRUonExists", benchHasEventExistsTest)
	b.Run("LRUonAbsent", benchHasEventAbsentTest)

	testStore = simpleStore
	b.Run("LRUoffExists", benchHasEventExistsTest)
	b.Run("LRUoffAbsent", benchHasEventAbsentTest)
}

func benchHasEventExistsTest(b *testing.B) {
	testEvent := &inter.Event{}

	testStore.SetEvent(testEvent)

	hev := testEvent.Hash()
	for i := 0; i < b.N; i++ {
		if !testStore.HasEvent(hev) {
			b.Fatalf("Not exists saved event\n")
		}
	}
}

func benchHasEventAbsentTest(b *testing.B) {
	testEvent := &inter.Event{}

	testStore.DeleteEvent(testEvent.Epoch, testEvent.Hash())

	hev := testEvent.Hash()
	for i := 0; i < b.N; i++ {
		if testStore.HasEvent(hev) {
			b.Fatalf("Exists absent event\n")
		}
	}
}
