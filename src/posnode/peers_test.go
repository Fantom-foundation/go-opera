package posnode

import (
	"testing"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/stretchr/testify/assert"
)

func TestPeerReadyForReq(t *testing.T) {
	// node 1
	store1 := NewMemStore()
	node1 := NewForTests("server1", store1, nil)

	// node 2
	store2 := NewMemStore()
	node2 := NewForTests("server2", store2, nil)

	// node1 got event1
	store1.BootstrapPeers(node2.AsPeer())
	node1.initPeers()

	t.Run("need host", func(t *testing.T) {
		assert := assert.New(t)

		attr := node1.peers.attrOf(node2.ID)
		attr.LastHost = "old_host"

		assert.Equal(node1.PeerReadyForReq(node2.ID, "new_host"), true)
	})

	t.Run("last success", func(t *testing.T) {
		assert := assert.New(t)

		attr := node1.peers.attrOf(node2.ID)
		attr.LastHost = "old_host"
		attr.LastFail = time.Now().Add(-time.Hour)
		attr.LastSuccess = time.Now()

		assert.Equal(node1.PeerReadyForReq(node2.ID, "old_host"), true)
	})

	t.Run("last fail timeouted", func(t *testing.T) {
		assert := assert.New(t)

		attr := node1.peers.attrOf(node2.ID)
		attr.LastHost = "old_host"
		attr.LastFail = time.Now().Add(-2 * node1.conf.DiscoveryTimeout)
		attr.LastSuccess = time.Now().Add(-node1.conf.DiscoveryTimeout)

		assert.Equal(node1.PeerReadyForReq(node2.ID, "old_host"), true)
	})
}

func TestPeerUnknown(t *testing.T) {
	// node 1
	store1 := NewMemStore()
	node1 := NewForTests("server1", store1, nil)

	// node 2
	store2 := NewMemStore()
	node2 := NewForTests("server2", store2, nil)

	// node1 got event1
	store1.BootstrapPeers(node2.AsPeer())
	node1.initPeers()

	t.Run("last success", func(t *testing.T) {
		assert := assert.New(t)

		attr := node1.peers.attrOf(node2.ID)
		attr.LastSuccess = time.Now().Truncate(2 * node1.conf.DiscoveryTimeout)

		assert.Equal(node1.PeerUnknown(&node2.ID), false)
	})

	t.Run("peer known", func(t *testing.T) {
		assert := assert.New(t)

		attr := node1.peers.attrOf(node2.ID)
		attr.LastSuccess = time.Now()

		assert.Equal(node1.PeerUnknown(&node2.ID), false)
	})

	t.Run("peer unknown", func(t *testing.T) {
		assert := assert.New(t)

		unknown := hash.HexToPeer("unknown")

		assert.Equal(node1.PeerUnknown(&unknown), true)
	})

	t.Run("nil peer", func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(node1.PeerUnknown(nil), true)
	})
}
