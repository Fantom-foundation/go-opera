package posnode

//NOTE: mockgen does not work properly out of GOPATH
//go:generate mockgen -package=posnode -source=consensus.go -destination=mock_test.go Consensus

import (
	"crypto/ecdsa"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/network"
)

func TestNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consensus := NewMockConsensus(ctrl)

	store := NewMemStore()

	key, err := crypto.GenerateECDSAKey()
	if err != nil {
		t.Fatal(err)
	}

	// TODO: network.FakeConnect instead of nil for tests.
	n := NewWithName(key, store, consensus, nil, "node001")
	defer n.Shutdown()

	// TODO: use network.FakeListener("") for tests.
	listener := network.TcpListener("")
	n.StartService(listener)
	defer n.StopService()
	t.Logf("node listen at %v", listener.Addr())

	n.StartDiscovery()
	defer n.StopDiscovery()

	n.StartGossip(4)
	defer n.StopGossip()
	<-time.After(5 * time.Second)
}

/*
 * Utils:
 */

// New creates node.
// TODO: use common.NodeNameDict instead of name after PR #161
func NewWithName(key *ecdsa.PrivateKey, s *Store, c Consensus, peerDialer Dialer, name string) *Node {
	id := CalcNodeID(&key.PublicKey)
	GetLogger(id, name)

	return New(key, s, c, peerDialer)
}
