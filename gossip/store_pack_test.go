package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStoreGetPackInfo(t *testing.T) {
	store := lruStore

	expect := &PackInfo{}
	expect.Index = idx.Pack(1)

	store.SetPackInfo(1, expect.Index, *expect)
	got := store.GetPackInfo(1, expect.Index)

	assert.EqualValues(t, expect, got)
}

func BenchmarkReadPackInfo(b *testing.B) {
	testStore = lruStore
	b.Run("LRUon", benchReadPackInfo)

	testStore = simpleStore
	b.Run("LRUoff", benchReadPackInfo)
}

func benchReadPackInfo(b *testing.B) {
	expect := &PackInfo{}
	expect.Index = idx.Pack(1)

	if testStore.cache.PackInfos != nil {
		testStore.cache.PackInfos.Purge()
	}

	testStore.SetPackInfo(1, expect.Index, *expect)

	for i := 0; i < b.N; i++ {
		_ = testStore.GetPackInfo(1, expect.Index)
	}
}

func BenchmarkWritePackInfo(b *testing.B) {
	testStore = lruStore
	b.Run("LRUon", benchWritePackInfo)

	testStore = simpleStore
	b.Run("LRUoff", benchWritePackInfo)
}

func benchWritePackInfo(b *testing.B) {
	expect := &PackInfo{}
	expect.Index = idx.Pack(1)

	if testStore.cache.PackInfos != nil {
		testStore.cache.PackInfos.Purge()
	}

	for i := 0; i < b.N; i++ {
		testStore.SetPackInfo(1, expect.Index, *expect)
	}
}
