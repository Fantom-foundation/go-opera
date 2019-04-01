package posposet

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

func TestMemStoreEvents(t *testing.T) {
	store := NewMemStore()

	t.Run("NotExisting", func(t *testing.T) {
		assert := assert.New(t)

		h := hash.FakeEvent()
		e1 := store.GetEvent(h)
		assert.Nil(e1)
	})

	t.Run("Events", func(t *testing.T) {
		assert := assert.New(t)

		events := FakeFuzzingEvents()
		for _, e0 := range events {
			store.SetEvent(e0)
			e1 := store.GetEvent(e0.Hash())

			if !assert.Equal(e0.Hash(), e1.Hash()) {
				break
			}
			if !assert.Equal(e0, e1) {
				break
			}
		}
	})

	store.Close()
}

func TestIntToKey(t *testing.T) {
	assert := assert.New(t)

	tests := map[uint64][]byte{
		0x0:                {0x0, 0, 0, 0, 0, 0, 0, 0},
		0x102:              {0x2, 0x1, 0, 0, 0, 0, 0, 0},
		0xFFFFFFFFFFFFFFFF: {0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	}

	for n, key0 := range tests {
		key1 := intToKey(n)

		if !assert.Equal(key0, key1) {
			break
		}
	}
}
