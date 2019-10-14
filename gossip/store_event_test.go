package gossip

import (
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

func TestStoreEventsCache(t *testing.T) {
	logger.SetTestMode(t)

	expect := fakeEvent()
	store := cachedStore()
	store.SetEvent(expect)

	got := store.GetEvent(expect.Hash())
	assert.EqualValues(t, expect, got)
}

func BenchmarkStoreGetEvent(b *testing.B) {
	logger.SetTestMode(b)

	b.Run("cache on", func(b *testing.B) {
		benchStoreGetEvent(b, cachedStore())
	})
	b.Run("cache off", func(b *testing.B) {
		benchStoreGetEvent(b, nonCachedStore())
	})
}

func benchStoreGetEvent(b *testing.B, store *Store) {
	e := fakeEvent()

	store.SetEvent(e)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if store.GetEvent(e.Hash()) == nil {
			b.Fatal("invalid result")
		}
	}
}

func BenchmarkStoreSetEvent(b *testing.B) {
	logger.SetTestMode(b)

	b.Run("cache on", func(b *testing.B) {
		benchStoreSetEvent(b, cachedStore())
	})
	b.Run("cache off", func(b *testing.B) {
		benchStoreSetEvent(b, nonCachedStore())
	})
}

func benchStoreSetEvent(b *testing.B, store *Store) {
	e := fakeEvent()

	for i := 0; i < b.N; i++ {
		store.SetEvent(e)
	}
}

func BenchmarkStoreHasEvent(b *testing.B) {
	logger.SetTestMode(b)

	b.Run("cache on and exists", func(b *testing.B) {
		benchStoreHasEvent(b, cachedStore(), true)
	})
	b.Run("cache off and exists", func(b *testing.B) {
		benchStoreHasEvent(b, nonCachedStore(), true)
	})
	b.Run("cache on and absent", func(b *testing.B) {
		benchStoreHasEvent(b, cachedStore(), false)
	})
	b.Run("cache off and absent", func(b *testing.B) {
		benchStoreHasEvent(b, nonCachedStore(), false)
	})
}

func benchStoreHasEvent(b *testing.B, store *Store, exists bool) {
	e := fakeEvent()
	h := e.Hash()

	if exists {
		store.SetEvent(e)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if store.HasEvent(h) != exists {
			b.Fatal("invalid result")
		}
	}
}

func fakeEvent() *inter.Event {
	d := &inter.Event{
		EventHeader: inter.EventHeader{
			EventHeaderData: inter.EventHeaderData{
				Parents:     hash.Events{},
				Extra:       make([]byte, 0),
				ClaimedTime: inter.Timestamp(rand.Int63()),
			},
			Sig: make([]byte, 0),
		},
		Transactions: types.Transactions{},
	}

	return d
}
