package posnode

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
)

func TestGossip(t *testing.T) {
	assert := assert.New(t)

	peers := generatePeers(2)
	events := inter.FakeFuzzingEvents()

	// Node 1
	store1 := NewMemStore()
	store1.BootstrapPeers(peers[1:])

	node1 := NewForTests(peers[0].Host, store1, nil)
	defer node1.Shutdown()

	// Set base height
	node1.store_SetHeights(&api.KnownEvents{
		Lasts: map[string]uint64{
			peers[0].ID.Hex(): 1,
		}})

	// Set base hash for event
	node1.store.SetEventHash(peers[0].ID, 1, events[0].Hash())

	// Set base event
	node1.store.SetEvent(events[0])

	node1.StartServiceForTests()
	defer node1.StopService()

	// Node 2
	store2 := NewMemStore()
	store2.BootstrapPeers(peers[:1])

	node2 := NewForTests(peers[1].Host, store2, nil)
	defer node2.Shutdown()

	// Set base height
	node2.store_SetHeights(&api.KnownEvents{
		Lasts: map[string]uint64{
			peers[1].ID.Hex(): 1,
		}})

	// Set base hash for event
	node2.store.SetEventHash(peers[1].ID, 1, events[1].Hash())

	// Set base event
	node2.store.SetEvent(events[1])

	node2.StartServiceForTests()
	defer node2.StopService()

	// gossip
	node1.syncWithPeer()
	node2.syncWithPeer()

	// check heights
	heights1 := node1.store_GetHeights()
	heights2 := node2.store_GetHeights()

	assert.Equal(heights1, heights2, "heights after gossiping")

	// check events
	hash1 := events[0].Hash()
	firstN1 := node1.store.GetEvent(hash1)
	firstN2 := node2.store.GetEvent(hash1)

	assert.Equal(firstN1, firstN2, "first event from both nodes")

	hash2 := events[1].Hash()
	secondN1 := node1.store.GetEvent(hash2)
	secondN2 := node2.store.GetEvent(hash2)

	assert.Equal(secondN1, secondN2, "second event from both nodes")
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
