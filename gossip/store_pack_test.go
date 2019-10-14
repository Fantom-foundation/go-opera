package gossip

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

func TestStoreGetPackInfo(t *testing.T) {
	store := _lruStore()

	expect := &PackInfo{}
	expect.Index = idx.Pack(1)

	store.SetPackInfo(1, expect.Index, *expect)
	got := store.GetPackInfo(1, expect.Index)

	assert.EqualValues(t, expect, got)
}

func BenchmarkReadPackInfo(b *testing.B) {
	b.Run("LRU on", func(b *testing.B) {
		benchReadPackInfo(b, _lruStore())
	})
	b.Run("LRU off", func(b *testing.B) {
		benchReadPackInfo(b, _simpleStore())
	})
}

func benchReadPackInfo(b *testing.B, store *Store) {
	expect := &PackInfo{}
	expect.Index = idx.Pack(1)

	if store.cache.PackInfos != nil {
		store.cache.PackInfos.Purge()
	}

	store.SetPackInfo(1, expect.Index, *expect)

	for i := 0; i < b.N; i++ {
		_ = store.GetPackInfo(1, expect.Index)
	}
}

func BenchmarkWritePackInfo(b *testing.B) {
	b.Run("LRU on", func(b *testing.B) {
		benchWritePackInfo(b, _lruStore())
	})
	b.Run("LRU off", func(b *testing.B) {
		benchWritePackInfo(b, _simpleStore())
	})
}

func benchWritePackInfo(b *testing.B, store *Store) {
	expect := &PackInfo{}
	expect.Index = idx.Pack(1)

	if store.cache.PackInfos != nil {
		store.cache.PackInfos.Purge()
	}

	for i := 0; i < b.N; i++ {
		store.SetPackInfo(1, expect.Index, *expect)
	}
}
