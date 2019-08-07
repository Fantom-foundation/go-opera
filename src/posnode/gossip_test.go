package posnode

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

func TestGossip(t *testing.T) {
	logger.SetTestMode(t)

	const sf = idx.SuperFrame(0)

	// node 1
	store1 := NewMemStore()
	node1 := NewForTests("node1", store1, nil)
	node1.StartService()
	defer node1.Stop()

	// node 2
	store2 := NewMemStore()
	node2 := NewForTests("node2", store2, nil)
	node2.StartService()
	defer node2.Stop()

	// connect nodes to each other
	store1.BootstrapPeers(node2.AsPeer())
	node1.initPeers()
	store2.BootstrapPeers(node1.AsPeer())
	node2.initPeers()

	// set events
	node1.EmitEvent()
	node2.EmitEvent()

	t.Run("before", func(t *testing.T) {
		assertar := assert.New(t)

		events1 := node1.knownEvents(0)
		assertar.Equal(
			heights{
				node1.ID: interval{to: 1},
				node2.ID: interval{to: 0},
			},
			events1,
			"node1 knows their event only")

		events2 := node2.knownEvents(0)
		assertar.Equal(
			heights{
				node1.ID: interval{to: 0},
				node2.ID: interval{to: 1},
			},
			events2,
			"node2 knows their event only")
	})

	t.Run("after 1-2", func(t *testing.T) {
		assertar := assert.New(t)

		node1.syncWithPeer(node2.AsPeer())

		events1 := node1.knownEvents(0)
		assertar.Equal(
			heights{
				node1.ID: interval{to: 1},
				node2.ID: interval{to: 1},
			},
			events1,
			"node1 knows last event of node2")

		events2 := node2.knownEvents(0)
		assertar.Equal(
			heights{
				node1.ID: interval{to: 0},
				node2.ID: interval{to: 1},
			},
			events2,
			"node2 still knows their event only")

		e2 := node1.store.GetEventHash(node2.ID, sf, 1)
		assertar.NotNil(e2, "event of node2 is in db")
	})

	t.Run("after 2-1", func(t *testing.T) {
		assertar := assert.New(t)

		node2.syncWithPeer(node1.AsPeer())

		events1 := node1.knownEvents(0)
		assertar.Equal(
			heights{
				node1.ID: interval{to: 1},
				node2.ID: interval{to: 1},
			},
			events1,
			"node1 still knows event of node2")

		events2 := node2.knownEvents(0)
		assertar.Equal(
			heights{
				node1.ID: interval{to: 1},
				node2.ID: interval{to: 1},
			},
			events2,
			"node2 knows last event of node1")

		e1 := node2.store.GetEventHash(node1.ID, sf, 1)
		assertar.NotNil(e1, "event of node1 is in db")
	})

}

func TestMissingParents(t *testing.T) {
	logger.SetTestMode(t)

	const sf = idx.SuperFrame(0)

	// node 1
	store1 := NewMemStore()
	node1 := NewForTests("node1", store1, nil)
	node1.StartService()
	defer node1.Stop()

	// node 2
	store2 := NewMemStore()
	node2 := NewForTests("node2", store2, nil)
	node2.StartService()
	defer node2.Stop()

	// connect nodes to each other
	store1.BootstrapPeers(node2.AsPeer())
	node1.initPeers()

	store2.BootstrapPeers(node1.AsPeer())
	node2.initPeers()

	node1.emitEvent()
	node1.emitEvent()
	node1.emitEvent()

	t.Run("before sync", func(t *testing.T) {
		assertar := assert.New(t)

		events1 := node1.knownEvents(0)
		assertar.Equal(
			heights{
				node1.ID: interval{to: 3},
				node2.ID: interval{to: 0},
			},
			events1,
			"node1 knows their event only")

		events2 := node2.knownEvents(0)
		assertar.Equal(
			heights{
				node1.ID: interval{to: 0},
				node2.ID: interval{to: 0},
			},
			events2,
			"node2 knows their event only")
	})

	// Note: we cannot use syncWithPeer here because we need custom iterator for missing some events.
	t.Run("sync", func(t *testing.T) {
		assertar := assert.New(t)

		peer := node1.AsPeer()
		client, _, _, err := node2.ConnectTo(peer)
		if err != nil {
			t.Fatal(err)
		}

		unknowns, err := node2.compareKnownEvents(client, peer, sf)
		if err != nil {
			t.Fatal(err)
		}
		if unknowns == nil {
			t.Fatal("unknowns is nil")
		}

		parents := hash.Events{}

		// sync with node1 but get only half events
		for creator, height := range unknowns {
			req := &api.EventRequest{
				PeerID: creator.Hex(),
			}
			// skipping the first event.
			for i := height.from + 1; i <= height.to; i++ {
				req.Seq = uint64(i)

				event, err := node2.downloadEvent(client, peer, req)
				if err != nil {
					t.Fatal(err)
				}
				if event == nil {
					t.Fatal("event is nil")
				}

				parents.Add(event.Parents...)
			}
		}

		// download missings
		node2.checkParents(client, peer, parents)

		events1 := node1.knownEvents(0)
		assertar.Equal(
			heights{
				node1.ID: interval{to: 3},
				node2.ID: interval{to: 0},
			},
			events1,
			"node1 still knows their event only")

		events2 := node2.knownEvents(0)
		assertar.Equal(
			heights{
				node1.ID: interval{to: 3},
				node2.ID: interval{to: 0},
			},
			events2,
			"node2 knows last event of node1")

		e1 := node2.store.GetEventHash(node1.ID, sf, 1)
		assertar.NotNil(e1, "event of node1 is in db")

		e2 := node2.store.GetEventHash(node1.ID, sf, 2)
		assertar.NotNil(e2, "event of node1 is in db")

		e3 := node2.store.GetEventHash(node1.ID, sf, 3)
		assertar.NotNil(e3, "event of node1 is in db")
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
	defer node.Stop()

	store.BootstrapPeers(peer1)
	node.initPeers()

	t.Run("First selection after bootstrap", func(t *testing.T) {
		assertar := assert.New(t)

		peer := node.NextForGossip()
		defer node.FreePeer(peer)

		assertar.Equal(
			peer1.Host,
			peer.Host,
			"node1 should select first top node without sort")
	})

	t.Run("Select last successful peer", func(t *testing.T) {
		assertar := assert.New(t)

		node.ConnectOK(peer1)
		store.SetPeer(peer2)
		node.ConnectOK(peer2)

		peer := node.NextForGossip()
		defer node.FreePeer(peer)

		assertar.Equal(
			peer2,
			peer,
			"node1 should select peer2 as first successful peer")
	})

	t.Run("Select last but one successful peer", func(t *testing.T) {
		assertar := assert.New(t)

		node.ConnectFail(peer2, nil)

		peer := node.NextForGossip()
		defer node.FreePeer(peer)

		assertar.Equal(
			peer1,
			peer,
			"should select peer1 as last successful peer")
	})

	t.Run("If all connection was failed -> select with earliest timestamp", func(t *testing.T) {
		assertar := assert.New(t)

		node.ConnectOK(peer1)
		node.ConnectFail(peer1, nil)
		node.ConnectFail(peer2, nil)

		peer := node.NextForGossip()
		defer node.FreePeer(peer)

		assertar.Equal(
			peer1,
			peer,
			"should select peer1 as last failed peer")
	})

	t.Run("If all connection was successful -> select first in top without sort", func(t *testing.T) {
		assertar := assert.New(t)

		node.ConnectOK(peer1)
		node.ConnectOK(peer2)

		peer := node.NextForGossip()
		defer node.FreePeer(peer)

		assertar.Equal(
			peer1,
			peer,
			"should select peer1 as first not busy in top peer")
	})
}
