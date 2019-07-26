package posnode

import (
	"github.com/Fantom-foundation/go-lachesis/src/cryptoaddr"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/fortytw2/leaktest"
)

func TestDiscoveryPeer(t *testing.T) {
	logger.SetTestMode(t)

	defer leaktest.CheckTimeout(t, time.Second)()

	// node 1
	store1 := NewMemStore()
	node1 := NewForTests("node1", store1, nil)
	node1.StartService()
	defer node1.Stop()

	// node 2
	store2 := NewMemStore()
	node2 := NewForTests("node2", store2, nil)
	node2.conf.ConnectTimeout = time.Millisecond * 100
	defer node2.Stop()

	// connect node2 to node1
	store2.BootstrapPeers(node1.AsPeer())
	node2.initPeers()

	t.Run("ask for unknown", func(t *testing.T) {
		assertar := assert.New(t)

		unknown := hash.FakePeer()
		node2.AskPeerInfo(node1.host, &unknown)

		peer := store2.GetPeer(unknown)
		assertar.Nil(peer)
	})

	t.Run("ask for himself", func(t *testing.T) {
		assertar := assert.New(t)

		unknown := node1.ID
		node2.AskPeerInfo(node1.host, &unknown)

		peer := store2.GetPeer(unknown)
		assertar.Equal(node1.AsPeer(), peer)
	})

	t.Run("ask for known unreachable", func(t *testing.T) {
		assertar := assert.New(t)

		known := FakePeer("unreachable")
		store1.SetPeer(known)

		node2.AskPeerInfo(node1.host, &known.ID)

		peer := store2.GetPeer(known.ID)
		assertar.Nil(peer)
	})

	t.Run("ask for known invalid", func(t *testing.T) {
		assertar := assert.New(t)

		known := InvalidPeer("invalid")
		store1.SetPeer(known)

		node2.AskPeerInfo(node1.host, &known.ID)

		peer := store2.GetPeer(known.ID)
		assertar.Nil(peer)
	})
}

func TestDiscoveryHost(t *testing.T) {
	logger.SetTestMode(t)

	defer leaktest.CheckTimeout(t, time.Second)()

	assertar := assert.New(t)

	// node 1
	store1 := NewMemStore()
	node1 := NewForTests("node1", store1, nil)
	node1.StartService()
	node1.initPeers()
	defer node1.Stop()

	// node 2
	store2 := NewMemStore()
	node2 := NewForTests("node2", store2, nil)
	node2.StartService()
	node2.StartDiscovery()
	defer node2.Stop()

	node1.AskPeerInfo(node2.Host(), nil)

	assertar.Equal(
		node2.ID,
		node1.peers.Snapshot()[0])

	select {
	case <-nodeDiscoveryFinish(node2):
	case <-time.After(time.Second):
	}
	assertar.Equal(
		node1.ID,
		node2.peers.Snapshot()[0])
}

/*
 * Utils:
 */

// FakePeer returns fake peer info.
func FakePeer(host string) *Peer {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}

	return &Peer{
		ID:     cryptoaddr.AddressOf(key.Public()),
		Host:   host,
	}
}

// InvalidPeer returns invalid peer info.
func InvalidPeer(host string) *Peer {
	peer := FakePeer(host)
	peer.ID = hash.FakePeer()
	return peer
}

func nodeDiscoveryFinish(n *Node) chan struct{} {
	done := make(chan struct{})
	go func() {
		for len(n.peers.Snapshot()) < 1 {
			time.Sleep(time.Second / 4)
		}
		close(done)
	}()
	return done
}
