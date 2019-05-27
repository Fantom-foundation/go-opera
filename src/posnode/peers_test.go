package posnode

import (
	"testing"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/stretchr/testify/assert"
)

func TestPeerReadyForReq(t *testing.T) {
	store := NewMemStore()
	node := NewForTests("node01", store, nil)
	node.initPeers()

	t.Run("new host", func(t *testing.T) {
		assert := assert.New(t)

		assert.True(node.PeerReadyForReq("new_host"))
	})

	t.Run("last success", func(t *testing.T) {
		assert := assert.New(t)

		host := &hostAttr{
			Name:        "success",
			LastSuccess: time.Now(),
			LastFail:    time.Now().Add(-node.conf.DiscoveryTimeout),
		}
		node.peers.hosts[host.Name] = host

		assert.True(node.PeerReadyForReq(host.Name))
	})

	t.Run("last fail timeouted", func(t *testing.T) {
		assert := assert.New(t)

		host := &hostAttr{
			Name:        "fail",
			LastSuccess: time.Now().Add(-node.conf.DiscoveryTimeout),
			LastFail:    time.Now().Add(-2 * node.conf.DiscoveryTimeout),
		}
		node.peers.hosts[host.Name] = host

		assert.True(node.PeerReadyForReq(host.Name))
	})
}

func TestPeerUnknown(t *testing.T) {
	store := NewMemStore()
	node := NewForTests("node02", store, nil)
	node.initPeers()

	t.Run("last success", func(t *testing.T) {
		assert := assert.New(t)

		peer := &peerAttr{
			ID: hash.FakePeer(),
			Host: &hostAttr{
				LastSuccess: time.Now().Truncate(node.conf.DiscoveryTimeout),
			},
		}
		node.peers.ids[peer.ID] = peer

		assert.False(node.PeerUnknown(&peer.ID))
	})

	t.Run("peer known", func(t *testing.T) {
		assert := assert.New(t)

		peer := &peerAttr{
			ID: hash.FakePeer(),
			Host: &hostAttr{
				LastSuccess: time.Now(),
			},
		}
		node.peers.ids[peer.ID] = peer

		assert.False(node.PeerUnknown(&peer.ID))
	})

	t.Run("peer unknown", func(t *testing.T) {
		assert := assert.New(t)

		unknown := hash.FakePeer()
		assert.True(node.PeerUnknown(&unknown))
	})

	t.Run("nil peer", func(t *testing.T) {
		assert := assert.New(t)

		assert.True(node.PeerUnknown(nil))
	})
}

// TODO: refactor tests
func TestCleanPeers(t *testing.T) {
	// peers
	peer1 := NewForTests("node1", NewMemStore(), nil).AsPeer()
	peer2 := NewForTests("node2", NewMemStore(), nil).AsPeer()

	// node
	store := NewMemStore()
	node := NewForTests("node", store, nil)
	node.StartService()
	defer node.StopService()

	peers := []*Peer{peer1, peer2}

	store.BootstrapPeers(peers...)
	node.initPeers()

	t.Run("leave hosts as is", func(t *testing.T) {
		assert := assert.New(t)

		node.PeerReadyForReq(peer1.Host)
		node.PeerReadyForReq(peer2.Host)

		node.trimHosts(2, 2)

		assert.Equal(len(node.peers.hosts), len(peers))
	})

	t.Run("clean expired hosts", func(t *testing.T) {
		assert := assert.New(t)

		// We already have info about 2 hosts but limit is 1
		// Clean all expired hosts
		node.trimHosts(1, 1)

		assert.Equal(len(node.peers.hosts), 1)
	})

	t.Run("clean extra hosts", func(t *testing.T) {
		assert := assert.New(t)

		node.PeerReadyForReq(peer1.Host)
		node.PeerReadyForReq(peer2.Host)

		node.ConnectOK(peer1)
		node.ConnectOK(peer2)

		// Clean extra
		node.trimHosts(1, 1)

		assert.Equal(len(node.peers.hosts), 1)
	})
}
