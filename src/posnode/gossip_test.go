package posnode

import (
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

func TestGossip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	peers := generatePeers(2)

	// Node 1
	consensus1 := NewMockConsensus(ctrl)
	store1 := NewMemStore()
	store1.BootstrapPeers(peers[1:])

	node1 := NewForTests(peers[0].Host, store1, consensus1)
	defer node1.Shutdown()

	node1.store_SetHeights(&wire.KnownEvents{
		Lasts: map[string]uint64{
			peers[0].ID.Hex(): 3,
		}})

	node1.StartServiceForTests()
	defer node1.StopService()

	node1.StartGossip(2)
	defer node1.StopGossip()

	// Node 2
	consensus2 := NewMockConsensus(ctrl)
	store2 := NewMemStore()
	store2.BootstrapPeers(peers[:1])

	node2 := NewForTests(peers[1].Host, store2, consensus2)
	defer node2.Shutdown()

	// Init data for node2
	node2.store_SetHeights(&wire.KnownEvents{
		Lasts: map[string]uint64{
			peers[1].ID.Hex(): 5,
		}})

	node2.StartServiceForTests()
	defer node2.StopService()

	node2.StartGossip(2)
	defer node2.StopGossip()

	<-time.After(3 * time.Second) // TODO: refactor whole test and remove it

	// Check
	heights1 := node1.store_GetHeights()
	heights2 := node2.store_GetHeights()

	for pID, height := range heights1.Lasts {
		if heights2.Lasts[pID] != height {
			t.Fatal("Error: Incorrect gossip data")
		}
	}
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
