package posnode

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

func TestEmit(t *testing.T) {
	// node 1
	store1 := NewMemStore()
	node1 := NewForTests("emitter1", store1, nil)
	node1.initParents()

	// node 2
	store2 := NewMemStore()
	node2 := NewForTests("emitter2", store2, nil)
	node2.initParents()

	events := make([]*inter.Event, 4)

	t.Run("very 1st event", func(t *testing.T) {
		assert := assert.New(t)
		// node1 has no candidates to parent
		tx := []byte("12345")
		node1.AddExternalTxn(tx)

		events[0] = node1.EmitEvent()

		assert.Equal(
			uint64(1),
			events[0].Index)
		assert.Equal(
			inter.Timestamp(1),
			events[0].LamportTime)
		assert.Equal(
			hash.NewEvents(hash.ZeroEvent),
			events[0].Parents)
		assert.Equal(
			[][]byte{tx},
			events[0].ExternalTransactions)
	})

	t.Run("1st event", func(t *testing.T) {
		assert := assert.New(t)
		// node2 got event0
		node2.saveNewEvent(events[0])

		events[1] = node2.EmitEvent()

		assert.Equal(
			uint64(1),
			events[1].Index)
		assert.Equal(
			inter.Timestamp(2),
			events[1].LamportTime)
		assert.Equal(
			hash.NewEvents(hash.ZeroEvent, events[0].Hash()),
			events[1].Parents)
	})

	t.Run("2nd event", func(t *testing.T) {
		assert := assert.New(t)
		// node1 got event1
		node1.saveNewEvent(events[1])

		events[2] = node1.EmitEvent()

		assert.Equal(
			uint64(2),
			events[2].Index)
		assert.Equal(
			inter.Timestamp(3),
			events[2].LamportTime)
		assert.Equal(
			hash.NewEvents(events[0].Hash(), events[1].Hash()),
			events[2].Parents)
	})

	t.Run("3rd event", func(t *testing.T) {
		assert := assert.New(t)
		// node1 has no new parents
		// TODO: what about skip event creation?

		events[3] = node1.EmitEvent()

		assert.Equal(
			uint64(3),
			events[3].Index)
		assert.Equal(
			inter.Timestamp(4),
			events[3].LamportTime)
		assert.Equal(
			hash.NewEvents(events[2].Hash()),
			events[3].Parents)
	})
}
