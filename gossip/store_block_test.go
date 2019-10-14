package gossip

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

func TestStoreGetBlock(t *testing.T) {
	store := fakeLruStore()

	expect := &inter.Block{}
	expect.Time = inter.Timestamp(rand.Int63())
	expect.Index = idx.Block(1)

	store.SetBlock(expect)
	got := store.GetBlock(1)

	assert.EqualValues(t, expect, got)
}

func BenchmarkReadBlock(b *testing.B) {
	b.Run("LRU on", func(b *testing.B) {
		benchReadBlock(b, fakeLruStore())
	})
	b.Run("LRU off", func(b *testing.B) {
		benchReadBlock(b, fakeSimpleStore())
	})
}

func benchReadBlock(b *testing.B, store *Store) {
	expect := &inter.Block{}
	expect.Time = inter.Timestamp(rand.Int63())
	expect.Index = idx.Block(1)

	if store.cache.Blocks != nil {
		store.cache.Blocks.Purge()
	}

	store.SetBlock(expect)

	for i := 0; i < b.N; i++ {
		_ = store.GetBlock(1)
	}
}

func BenchmarkWriteBlock(b *testing.B) {
	b.Run("LRU on", func(b *testing.B) {
		benchWriteBlock(b, fakeLruStore())
	})
	b.Run("LRU off", func(b *testing.B) {
		benchWriteBlock(b, fakeSimpleStore())
	})
}

func benchWriteBlock(b *testing.B, store *Store) {
	expect := &inter.Block{}
	expect.Time = inter.Timestamp(rand.Int63())
	expect.Index = idx.Block(1)

	if store.cache.Blocks != nil {
		store.cache.Blocks.Purge()
	}

	for i := 0; i < b.N; i++ {
		store.SetBlock(expect)
	}
}
