package gossip

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

func TestStoreGetPackInfo(t *testing.T) {
	logger.SetTestMode(t)

	epoch, expect := fakePackInfo()
	store := cachedStore()
	store.SetPackInfo(epoch, expect.Index, *expect)

	got := store.GetPackInfo(epoch, expect.Index)
	assert.EqualValues(t, expect, got)
}

func BenchmarkStoreGetPackInfo(b *testing.B) {
	logger.SetTestMode(b)

	b.Run("cache on", func(b *testing.B) {
		benchStoreGetPackInfo(b, cachedStore())
	})
	b.Run("cache off", func(b *testing.B) {
		benchStoreGetPackInfo(b, nonCachedStore())
	})
}

func benchStoreGetPackInfo(b *testing.B, store *Store) {
	epoch, pack := fakePackInfo()

	store.SetPackInfo(epoch, pack.Index, *pack)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if store.GetPackInfo(epoch, pack.Index) == nil {
			b.Fatal("invalid result")
		}
	}
}

func BenchmarkStoreSetPackInfo(b *testing.B) {
	logger.SetTestMode(b)

	b.Run("cache on", func(b *testing.B) {
		benchStoreSetPackInfo(b, cachedStore())
	})
	b.Run("cache off", func(b *testing.B) {
		benchStoreSetPackInfo(b, nonCachedStore())
	})
}

func benchStoreSetPackInfo(b *testing.B, store *Store) {
	epoch, pack := fakePackInfo()

	for i := 0; i < b.N; i++ {
		store.SetPackInfo(epoch, pack.Index, *pack)
	}
}

func fakePackInfo() (idx.Epoch, *PackInfo) {
	return idx.Epoch(1),
		&PackInfo{
			Index: idx.Pack(1),
		}
}
