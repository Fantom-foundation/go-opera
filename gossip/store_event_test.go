package gossip

/*
	Benchmarks for store Events with LRU and without
*/

import (
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
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

	expect := &inter.Event{}
	expect.ClaimedTime = inter.Timestamp(rand.Int63())

	store.SetEvent(expect)
	got := store.GetEvent(expect.Hash())

	assert.EqualValues(t, expect, got)
}

func BenchmarkReadEvent(b *testing.B) {
	testStore = lruStore
	b.Run("LRUon", benchReadEventTest)

	testStore = simpleStore
	b.Run("LRUoff", benchReadEventTest)
}

func benchReadEventTest(b *testing.B) {
	expect := _createTestEvent()
	if testStore.cache.Events != nil {
		testStore.cache.Events.Purge()
	}

	testStore.SetEvent(expect)

	for i := 0; i < b.N; i++ {
		_ = testStore.GetEvent(expect.Hash())
	}
}

func BenchmarkWriteEvent(b *testing.B) {
	testStore = lruStore
	b.Run("LRUon", benchWriteEventTest)

	testStore = simpleStore
	b.Run("LRUoff", benchWriteEventTest)
}

func benchWriteEventTest(b *testing.B) {
	expect := &inter.Event{}

	for i := 0; i < b.N; i++ {
		testStore.SetEvent(expect)
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
	expect := &inter.Event{}

	testStore.SetEvent(expect)

	hev := expect.Hash()
	for i := 0; i < b.N; i++ {
		_ = testStore.HasEvent(hev)
	}
}

func benchHasEventAbsentTest(b *testing.B) {
	expect := &inter.Event{}

	testStore.DeleteEvent(expect.Epoch, expect.Hash())

	hev := expect.Hash()
	for i := 0; i < b.N; i++ {
		_ = testStore.HasEvent(hev)
	}
}

func _createTestEvent() *inter.Event {
	d := &inter.Event{
		EventHeader:  inter.EventHeader{
			EventHeaderData: inter.EventHeaderData{
				Parents:hash.Events{},
				Extra: make([]byte, 0),
			},
			Sig: make([]byte, 0),
		},
		Transactions: types.Transactions{},
	}

	return d
}
