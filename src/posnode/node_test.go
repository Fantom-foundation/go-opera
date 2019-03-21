package posnode

//NOTE: mockgen does not work properly out of GOPATH
//go:generate mockgen -package=posnode -source=consensus.go -destination=mock_test.go Consensus

import (
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

	key, err := crypto.GenerateECDSAKey()
	if err != nil {
		t.Fatal(err)
	}

	n := New(key, consensus, nil) // TODO: network.FakeConnect instead of nil for tests.
	defer n.Shutdown()

	// TODO: use network.FakeListener("") for tests.
	listener := network.TcpListener("")
	n.StartService(listener)
	defer n.StopService()
	t.Logf("node listen at %v", listener.Addr())

	n.StartGossip(4)
	defer n.StopGossip()
	<-time.After(5 * time.Second)
}
