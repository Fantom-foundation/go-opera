package posnode

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

func TestGossip(t *testing.T) {
	// node 1
	store1 := NewMemStore()
	node1 := NewForTests("node1", store1, nil)
	defer node1.Shutdown()
	node1.StartServiceForTests()
	defer node1.StopService()

	// node 2
	store2 := NewMemStore()
	node2 := NewForTests("node2", store2, nil)
	defer node2.Shutdown()
	node2.StartServiceForTests()
	defer node1.StopService()

	// connect nodes to each other
	store1.BootstrapPeers(&Peer{
		ID:     node2.ID,
		PubKey: node2.pub,
		Host:   node2.host,
	})
	node1.initPeers()
	store2.BootstrapPeers(&Peer{
		ID:     node1.ID,
		PubKey: node1.pub,
		Host:   node1.host,
	})
	node2.initPeers()

	// set events
	// TODO: replace with self-generated events
	node1.SaveNewEvent(&inter.Event{
		Index:   1,
		Creator: node1.ID,
	})
	node2.SaveNewEvent(&inter.Event{
		Index:   1,
		Creator: node2.ID,
	})

	t.Run("before", func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(
			map[string]uint64{
				node1.ID.Hex(): 1,
				node2.ID.Hex(): 0,
			},
			node1.knownEvents().Lasts,
			"node1 knows their event only")

		assert.Equal(
			map[string]uint64{
				node1.ID.Hex(): 0,
				node2.ID.Hex(): 1,
			},
			node2.knownEvents().Lasts,
			"node2 knows their event only")
	})

	t.Run("after 1-2", func(t *testing.T) {
		assert := assert.New(t)
		node1.syncWithPeer()

		assert.Equal(
			map[string]uint64{
				node1.ID.Hex(): 1,
				node2.ID.Hex(): 1,
			},
			node1.knownEvents().Lasts,
			"node1 knows last event of node2")

		assert.Equal(
			map[string]uint64{
				node1.ID.Hex(): 0,
				node2.ID.Hex(): 1,
			},
			node2.knownEvents().Lasts,
			"node2 still knows their event only")

		e2 := node1.store.GetEventHash(node2.ID, 1)
		assert.NotNil(e2, "event of node2 is in db")
	})

	t.Run("after 2-1", func(t *testing.T) {
		assert := assert.New(t)
		node2.syncWithPeer()

		assert.Equal(
			map[string]uint64{
				node1.ID.Hex(): 1,
				node2.ID.Hex(): 1,
			},
			node1.knownEvents().Lasts,
			"node1 still knows event of node2")

		assert.Equal(
			map[string]uint64{
				node1.ID.Hex(): 1,
				node2.ID.Hex(): 1,
			},
			node2.knownEvents().Lasts,
			"node2 knows last event of node1")

		e1 := node2.store.GetEventHash(node1.ID, 1)
		assert.NotNil(e1, "event of node1 is in db")
	})
}
