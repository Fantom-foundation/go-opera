package posnode

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

func TestGossip(t *testing.T) {
	// node 1
	store1 := NewMemStore()
	node1 := NewForTests("node1", store1, nil)
	node1.StartService()
	defer node1.StopService()

	// node 2
	store2 := NewMemStore()
	node2 := NewForTests("node2", store2, nil)
	node2.StartService()
	defer node2.StopService()

	// connect nodes to each other
	store1.BootstrapPeers(node2.AsPeer())
	node1.initPeers()
	store2.BootstrapPeers(node1.AsPeer())
	node2.initPeers()

	// set events
	// TODO: replace with self-generated events
	genEvents(node1, 1)
	genEvents(node2, 1)

	t.Run("before", func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(
			map[hash.Peer]uint64{
				node1.ID: 1,
				node2.ID: 0,
			},
			node1.knownEvents(),
			"node1 knows their event only")

		assert.Equal(
			map[hash.Peer]uint64{
				node1.ID: 0,
				node2.ID: 1,
			},
			node2.knownEvents(),
			"node2 knows their event only")
	})

	t.Run("after 1-2", func(t *testing.T) {
		assert := assert.New(t)
		node1.syncWithPeer(node2.AsPeer())

		assert.Equal(
			map[hash.Peer]uint64{
				node1.ID: 1,
				node2.ID: 1,
			},
			node1.knownEvents(),
			"node1 knows last event of node2")

		assert.Equal(
			map[hash.Peer]uint64{
				node1.ID: 0,
				node2.ID: 1,
			},
			node2.knownEvents(),
			"node2 still knows their event only")

		e2 := node1.store.GetEventHash(node2.ID, 1)
		assert.NotNil(e2, "event of node2 is in db")
	})

	t.Run("after 2-1", func(t *testing.T) {
		assert := assert.New(t)
		node2.syncWithPeer(node1.AsPeer())

		assert.Equal(
			map[hash.Peer]uint64{
				node1.ID: 1,
				node2.ID: 1,
			},
			node1.knownEvents(),
			"node1 still knows event of node2")

		assert.Equal(
			map[hash.Peer]uint64{
				node1.ID: 1,
				node2.ID: 1,
			},
			node2.knownEvents(),
			"node2 knows last event of node1")

		e1 := node2.store.GetEventHash(node1.ID, 1)
		assert.NotNil(e1, "event of node1 is in db")
	})

}

func TestMissingParents(t *testing.T) {
	// node 1
	store1 := NewMemStore()
	node1 := NewForTests("node1", store1, nil)
	node1.StartService()
	defer node1.StopService()

	// node 2
	store2 := NewMemStore()
	node2 := NewForTests("node2", store2, nil)
	node2.StartService()
	defer node1.StopService()

	// connect nodes to each other
	store1.BootstrapPeers(node2.AsPeer())
	node1.initPeers()

	store2.BootstrapPeers(node1.AsPeer())
	node2.initPeers()

	genEvents(node1, 2)

	t.Run("before", func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(
			map[hash.Peer]uint64{
				node1.ID: 2,
				node2.ID: 0,
			},
			node1.knownEvents(),
			"node1 knows their event only")

		assert.Equal(
			map[hash.Peer]uint64{
				node1.ID: 0,
				node2.ID: 0,
			},
			node2.knownEvents(),
			"node2 knows their event only")
	})

	t.Run("after 2-1", func(t *testing.T) {
		assert := assert.New(t)
		node2.syncWithPeer(node1.AsPeer())

		assert.Equal(
			map[hash.Peer]uint64{
				node1.ID: 2,
				node2.ID: 0,
			},
			node1.knownEvents(),
			"node1 still knows their event only")

		assert.Equal(
			map[hash.Peer]uint64{
				node1.ID: 2,
				node2.ID: 0,
			},
			node2.knownEvents(),
			"node2 knows last event of node1")

		e1 := node2.store.GetEventHash(node1.ID, 1)
		assert.NotNil(e1, "event of node1 is in db")

		e2 := node2.store.GetEventHash(node1.ID, 2)
		assert.NotNil(e2, "event of node1 is in db")
	})
}

func TestPeerPriority(t *testing.T) {
	// peers
	peer1 := NewForTests("node1", NewMemStore(), nil).AsPeer()
	peer2 := NewForTests("node2", NewMemStore(), nil).AsPeer()

	// node
	store := NewMemStore()
	node := NewForTests("node0", store, nil)
	node.StartService()
	defer node.StopService()

	store.BootstrapPeers(peer1)
	node.initPeers()

	t.Run("First selection after bootstrap", func(t *testing.T) {
		assert := assert.New(t)

		peer := node.NextForGossip()
		defer node.FreePeer(peer)

		assert.Equal(
			peer1.Host,
			peer.Host,
			"node1 should select first top node without sort")
	})

	t.Run("Select last successful peer", func(t *testing.T) {
		assert := assert.New(t)

		node.ConnectOK(peer1)
		store.SetPeer(peer2)
		node.ConnectOK(peer2)

		peer := node.NextForGossip()
		defer node.FreePeer(peer)

		assert.Equal(
			peer2,
			peer,
			"node1 should select peer2 as first successful peer")
	})

	t.Run("Select last but one successful peer", func(t *testing.T) {
		assert := assert.New(t)

		node.ConnectFail(peer2, nil)

		peer := node.NextForGossip()
		defer node.FreePeer(peer)

		assert.Equal(
			peer1,
			peer,
			"should select peer1 as last successful peer")
	})

	t.Run("If all connection was failed -> select with earliest timestamp", func(t *testing.T) {
		assert := assert.New(t)

		node.ConnectOK(peer1)
		node.ConnectFail(peer1, nil)
		node.ConnectFail(peer2, nil)

		peer := node.NextForGossip()
		defer node.FreePeer(peer)

		assert.Equal(
			peer1,
			peer,
			"should select peer1 as last failed peer")
	})

	t.Run("If all connection was successfull -> select first in top without sort", func(t *testing.T) {
		assert := assert.New(t)

		node.ConnectOK(peer1)
		node.ConnectOK(peer2)

		peer := node.NextForGossip()
		defer node.FreePeer(peer)

		assert.Equal(
			peer1,
			peer,
			"should select peer1 as first not busy in top peer")
	})
}

/*
 * Utils:
 */

func genEvents(n *Node, count uint64) {
	last := n.store.GetPeerHeight(n.ID)
	for i := last + 1; i <= last+count; i++ {
		e := &inter.Event{
			Index:   i,
			Creator: n.ID,
			Parents: hash.Events{},
		}
		if i == 1 {
			e.Parents.Add(hash.ZeroEvent)
		} else {
			e.Parents.Add(*n.store.GetEventHash(n.ID, i-1))
		}

		err := e.SignBy(n.key)
		if err != nil {
			panic(err)
		}

		n.saveNewEvent(e)
	}
}
