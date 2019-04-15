package posnode

import (
	"testing"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/stretchr/testify/assert"
)

func TestPeerReadyForReq(t *testing.T) {
	store := NewMemStore()
	node := NewForTests("node0", store, nil)
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
	node := NewForTests("node0", store, nil)
	node.initPeers()

	t.Run("last success", func(t *testing.T) {
		assert := assert.New(t)

		peer := &peerAttr{
			ID: hash.FakePeer(),
			Host: &hostAttr{
				LastSuccess: time.Now().Truncate(node.conf.DiscoveryTimeout),
			},
		}
		node.peers.peers[peer.ID] = peer

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
		node.peers.peers[peer.ID] = peer

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
