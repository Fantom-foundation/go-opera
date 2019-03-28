package posnode

import (
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

func TestGossip(t *testing.T) {

	return // TODO: fix deadlock

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Generate test peers
	peers := generatePeers(2)

	// Node 1
	consensus1 := NewMockConsensus(ctrl)
	store1 := NewMemStore()

	node1 := NewForTests("server.fake0", store1, consensus1)
	defer node1.Shutdown()

	// Init peer for node1
	peers1 := peers[1:]
	initPeers(node1, peers1)

	// Init data for node1
	heights := map[string]uint64{}
	heights[peers[0].NetAddr] = 3
	node1.store_SetHeights(&wire.KnownEvents{Lasts: heights})

	node1.StartServiceForTests()
	defer node1.StopService()

	node1.StartGossip(2)
	defer node1.StopGossip()

	// Node 2
	consensus2 := NewMockConsensus(ctrl)
	store2 := NewMemStore()

	node2 := NewForTests("server.fake1", store2, consensus2)
	defer node2.Shutdown()

	// Init peer for node2
	peers2 := peers[:1]
	initPeers(node2, peers2)

	// Init data for node2
	heights = map[string]uint64{}
	heights[peers[1].NetAddr] = 5
	node2.store_SetHeights(&wire.KnownEvents{Lasts: heights})

	node2.StartServiceForTests()
	defer node2.StopService()

	node2.StartGossip(2)
	defer node2.StopGossip()

	<-time.After(3 * time.Second)

	// Check
	heights1 := node1.store_GetHeights()
	heights2 := node2.store_GetHeights()

	for pID, height := range heights1.Lasts {
		if heights2.Lasts[pID] != height {
			t.Fatal("Error: Incorrect gossip data")
		}
	}
}

func initPeers(node *Node, peers []*Peer) {
	ids := []common.Address{}

	for _, peer := range peers {
		node.store.SetPeer(peer)

		ids = append(ids, peer.ID)
	}

	node.store.SetTopPeers(ids)
}

func generatePeers(count int) []*Peer {
	var peers []*Peer

	for i := 0; i < count; i++ {
		key, _ := crypto.GenerateECDSAKey()
		pubKey := key.PublicKey
		netAddr := "server.fake" + strconv.Itoa(i) + ":55555"

		peer := &Peer{
			ID:      CalcNodeID(&pubKey),
			PubKey:  &pubKey,
			NetAddr: netAddr,
		}

		peers = append(peers, peer)
	}

	return peers
}
