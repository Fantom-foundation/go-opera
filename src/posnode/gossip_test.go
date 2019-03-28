package posnode

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

func TestGossip(t *testing.T) {
	assert := assert.New(t)

	peers := generatePeers(2)

	// Node 1
	store1 := NewMemStore()
	store1.BootstrapPeers(peers[1:])

	node1 := NewForTests(peers[0].Host, store1, nil)
	defer node1.Shutdown()

	node1.store_SetHeights(&wire.KnownEvents{
		Lasts: map[string]uint64{
			peers[0].ID.Hex(): 3,
		}})

	node1.StartServiceForTests()
	defer node1.StopService()

	// Node 2
	store2 := NewMemStore()
	store2.BootstrapPeers(peers[:1])

	node2 := NewForTests(peers[1].Host, store2, nil)
	defer node2.Shutdown()

	// Init data for node2
	node2.store_SetHeights(&wire.KnownEvents{
		Lasts: map[string]uint64{
			peers[1].ID.Hex(): 5,
		}})

	node2.StartServiceForTests()
	defer node2.StopService()

	// gossip
	node1.syncWithPeer()
	node2.syncWithPeer()

	// check
	heights1 := node1.store_GetHeights()
	heights2 := node2.store_GetHeights()

	assert.Equal(heights1, heights2, "heights after gossiping")
}

func generatePeers(count int) []*Peer {
	var peers []*Peer

	for i := 0; i < count; i++ {
		key, _ := crypto.GenerateECDSAKey()

		peer := &Peer{
			ID:     CalcNodeID(&key.PublicKey),
			PubKey: &key.PublicKey,
			Host:   "server.fake." + strconv.Itoa(i),
		}

		peers = append(peers, peer)
	}

	return peers
}
