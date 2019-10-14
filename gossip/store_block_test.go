package gossip

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

func TestStoreGetBlock(t *testing.T) {
	store := lruStore

	expect := &inter.Block{}
	expect.Time = inter.Timestamp(rand.Int63())
	expect.Index = idx.Block(1)

	store.SetBlock(expect)
	got := store.GetBlock(1)

	assert.EqualValues(t, expect, got)
}

func BenchmarkReadBlock(b *testing.B) {
	testStore = lruStore
	b.Run("LRUon", benchReadBlock)

	testStore = simpleStore
	b.Run("LRUoff", benchReadBlock)
}

func benchReadBlock(b *testing.B) {
	expect := &inter.Block{}
	expect.Time = inter.Timestamp(rand.Int63())
	expect.Index = idx.Block(1)

	if testStore.cache.Blocks != nil {
		testStore.cache.Blocks.Purge()
	}

	testStore.SetBlock(expect)

	for i := 0; i < b.N; i++ {
		_ = testStore.GetBlock(1)
	}
}

func BenchmarkWriteBlock(b *testing.B) {
	testStore = lruStore
	b.Run("LRUon", benchWriteBlock)

	testStore = simpleStore
	b.Run("LRUoff", benchWriteBlock)
}

func benchWriteBlock(b *testing.B) {
	expect := &inter.Block{}
	expect.Time = inter.Timestamp(rand.Int63())
	expect.Index = idx.Block(1)

	if testStore.cache.Blocks != nil {
		testStore.cache.Blocks.Purge()
	}

	for i := 0; i < b.N; i++ {
		testStore.SetBlock(expect)
	}
}
