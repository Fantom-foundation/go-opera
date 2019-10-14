package gossip

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

func TestStoreGetEventHeader(t *testing.T) {
	logger.SetTestMode(t)

	store := cachedStore()
	expect := fakeEvent()
	h := expect.Hash()

	store.SetEventHeader(expect.Epoch, h, &expect.EventHeaderData)
	got := store.GetEventHeader(expect.Epoch, h)

	assert.EqualValues(t, &expect.EventHeaderData, got)
}

func BenchmarkStoreGetEventHeader(b *testing.B) {
	logger.SetTestMode(b)

	b.Run("cache on", func(b *testing.B) {
		benchStoreGetEventHeader(b, cachedStore())
	})
	b.Run("cache off", func(b *testing.B) {
		benchStoreGetEventHeader(b, nonCachedStore())
	})
}

func benchStoreGetEventHeader(b *testing.B, store *Store) {
	e := &inter.Event{}
	h := e.Hash()

	store.SetEventHeader(e.Epoch, h, &e.EventHeaderData)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if store.GetEventHeader(e.Epoch, h) == nil {
			b.Fatal("invalid result")
		}
	}
}

func BenchmarkStoreSetEventHeader(b *testing.B) {
	logger.SetTestMode(b)

	b.Run("cache on", func(b *testing.B) {
		benchStoreSetEventHeader(b, cachedStore())
	})
	b.Run("cache off", func(b *testing.B) {
		benchStoreSetEventHeader(b, nonCachedStore())
	})
}

func benchStoreSetEventHeader(b *testing.B, store *Store) {
	e := fakeEvent()
	h := e.Hash()

	for i := 0; i < b.N; i++ {
		store.SetEventHeader(e.Epoch, h, &e.EventHeaderData)
	}
}
