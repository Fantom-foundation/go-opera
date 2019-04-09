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
		node1.syncWithPeer()

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
		node2.syncWithPeer()

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
		node2.syncWithPeer()

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
	// Node
	store1 := NewMemStore()
	node1 := NewForTests("node1", store1, nil)
	node1.StartService()
	defer node1.StopService()

	// Peers
	node2 := NewForTests("node2", store1, nil)
	node3 := NewForTests("node3", store1, nil)

	// init peer list
	peer2 := &Peer{
		ID:     node2.ID,
		PubKey: node2.pub,
		Host:   node2.host,
	}

	peer3 := &Peer{
		ID:     node3.ID,
		PubKey: node3.pub,
		Host:   node3.host,
	}

	store1.BootstrapPeers(peer2)
	node1.initPeers()

	t.Run("First selection after bootstrap", func(t *testing.T) {
		assert := assert.New(t)

		peer := node1.NextForGossip()
		node1.FreePeer(peer)

		node1.ConnectOK(peer2)

		assert.Equal(
			peer2.Host,
			peer.Host,
			"node1 should select first top node without sort")
	})

	t.Run("Select last successful peer", func(t *testing.T) {
		assert := assert.New(t)

		// Add unknown peer for update top
		node1.ConnectOK(peer3)

		peer := node1.NextForGossip()
		node1.FreePeer(peer)

		assert.Equal(
			peer3.Host,
			peer.Host,
			"node1 should select peer3 as first successful node to connect after sort")
	})

	t.Run("Select last but one successful peer", func(t *testing.T) {
		assert := assert.New(t)

		node1.ConnectFail(peer3, nil)

		peer := node1.NextForGossip()
		node1.FreePeer(peer)

		assert.Equal(
			peer2.Host,
			peer.Host,
			"node1 should select peer2 as last successful node to connect after sort")
	})

	t.Run("If all connection was failed -> select with earliest timestamp", func(t *testing.T) {
		assert := assert.New(t)

		// Clean previous data for current case.
		node1.peers.attrs = make(map[hash.Peer]*peerAttrs)

		node1.ConnectFail(peer2, nil)
		node1.ConnectFail(peer3, nil)

		peer := node1.NextForGossip()
		node1.FreePeer(peer)

		assert.Equal(
			peer2.Host,
			peer.Host,
			"node1 should select peer2 as first failed node to connect after sort")
	})

	t.Run("If all connection was successfull -> select first in top without sort", func(t *testing.T) {
		assert := assert.New(t)

		node1.ConnectOK(peer2)
		node1.ConnectOK(peer3)

		peer := node1.NextForGossip()
		node1.FreePeer(peer)

		assert.Equal(
			peer2.Host,
			peer.Host,
			"node1 should select peer2 as first not busy in top node to connect after sort")
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
