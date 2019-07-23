package posnode

import (
	"testing"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/stretchr/testify/assert"
)

func TestPeerReadyForReq(t *testing.T) {
	logger.SetTestMode(t)

	store := NewMemStore()
	node := NewForTests("node01", store, nil)
	node.initPeers()
	defer node.Stop()

	t.Run("new host", func(t *testing.T) {
		assertar := assert.New(t)

		assertar.True(node.PeerReadyForReq("new_host"))
	})

	t.Run("last success", func(t *testing.T) {
		// TODO: don't skip
		t.Skip("see Node.PeerReadyForReq() todo")
		assertar := assert.New(t)

		host := &hostAttr{
			Name:        "success",
			LastSuccess: time.Now(),
			LastFail:    time.Now().Add(-node.conf.DiscoveryTimeout),
		}
		node.peers.hosts[host.Name] = host

		assertar.False(node.PeerReadyForReq(host.Name))
	})

	t.Run("last fail timeouted", func(t *testing.T) {
		assertar := assert.New(t)

		host := &hostAttr{
			Name:        "fail",
			LastSuccess: time.Now().Add(-node.conf.DiscoveryTimeout),
			LastFail:    time.Now().Add(-2 * node.conf.DiscoveryTimeout),
		}
		node.peers.hosts[host.Name] = host

		assertar.True(node.PeerReadyForReq(host.Name))
	})
}

func TestPeerUnknown(t *testing.T) {
	logger.SetTestMode(t)

	store := NewMemStore()
	node := NewForTests("node02", store, nil)
	node.initPeers()
	defer node.Stop()

	t.Run("last success", func(t *testing.T) {
		assertar := assert.New(t)

		peer := &peerAttr{
			ID: hash.FakePeer(),
			Host: &hostAttr{
				LastSuccess: time.Now().Truncate(node.conf.DiscoveryTimeout),
			},
		}
		node.peers.ids[peer.ID] = peer

		assertar.False(node.PeerUnknown(&peer.ID))
	})

	t.Run("peer known", func(t *testing.T) {
		assertar := assert.New(t)

		peer := &peerAttr{
			ID: hash.FakePeer(),
			Host: &hostAttr{
				LastSuccess: time.Now(),
			},
		}
		node.peers.ids[peer.ID] = peer

		assertar.False(node.PeerUnknown(&peer.ID))
	})

	t.Run("peer unknown", func(t *testing.T) {
		assertar := assert.New(t)

		unknown := hash.FakePeer()
		assertar.True(node.PeerUnknown(&unknown))
	})

	t.Run("nil peer", func(t *testing.T) {
		assertar := assert.New(t)

		assertar.True(node.PeerUnknown(nil))
	})
}

// TODO: refactor tests
func TestCleanPeers(t *testing.T) {
	logger.SetTestMode(t)

	// peers
	peer1 := NewForTests("node1", NewMemStore(), nil).AsPeer()
	peer2 := NewForTests("node2", NewMemStore(), nil).AsPeer()

	// node
	store := NewMemStore()
	node := NewForTests("node", store, nil)
	node.StartService()
	defer node.Stop()

	peers := []*Peer{peer1, peer2}

	store.BootstrapPeers(peers...)
	node.initPeers()

	t.Run("leave hosts as is", func(t *testing.T) {
		assertar := assert.New(t)

		node.PeerReadyForReq(peer1.Host)
		node.PeerReadyForReq(peer2.Host)

		node.trimHosts(2, 2)

		assertar.Equal(len(node.peers.hosts), len(peers))
	})

	t.Run("clean expired hosts", func(t *testing.T) {
		assertar := assert.New(t)

		// We already have info about 2 hosts but limit is 1
		// Clean all expired hosts
		node.trimHosts(1, 1)

		assertar.Equal(len(node.peers.hosts), 1)
	})

	t.Run("clean extra hosts", func(t *testing.T) {
		assertar := assert.New(t)

		node.PeerReadyForReq(peer1.Host)
		node.PeerReadyForReq(peer2.Host)

		node.ConnectOK(peer1)
		node.ConnectOK(peer2)

		// Clean extra
		node.trimHosts(1, 1)

		assertar.Equal(len(node.peers.hosts), 1)
	})
}
