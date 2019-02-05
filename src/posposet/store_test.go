package posposet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStoreEvents(t *testing.T) {
	store := NewMemStore()

	t.Run("NotExisting", func(t *testing.T) {
		assert := assert.New(t)

		hash := FakeEventHash()
		e1 := store.GetEvent(hash)
		assert.Nil(e1)
	})

	t.Run("Events", func(t *testing.T) {
		assert := assert.New(t)

		events := FakeEvents()
		for _, e0 := range events {
			store.SetEvent(e0)
			e1 := store.GetEvent(e0.Hash())
			assert.Equal(e0, e1)
		}
	})

	store.Close()
}
