package posnode

//NOTE: mockgen does not work properly out of GOPATH
//go:generate mockgen -package=posnode -source=consensus.go -destination=mock_test.go Consensus

import (
	"testing"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/network"
)

func ExampleNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consensus := NewMockConsensus(ctrl)

	store := NewMemStore()

	n := NewForTests("server.fake", store, consensus)
	defer n.Shutdown()

	n.StartServiceForTests()
	defer n.StopService()

	n.StartDiscovery()
	defer n.StopDiscovery()

	n.StartGossip(4)
	defer n.StopGossip()

	select {}
}

/*
 * Utils:
 */

// New creates node for tests.
func NewForTests(host string, s *Store, c Consensus) *Node {
	key, err := crypto.GenerateECDSAKey()
	if err != nil {
		panic(err)
	}

	dialer := network.FakeDialer(host)
	opts := grpc.WithContextDialer(dialer)

	return New(host, key, s, c, DefaultConfig(), opts)
}

// StartServiceForTests starts node service.
// It should be called once.
func (n *Node) StartServiceForTests() {
	if n.server != nil {
		return
	}
	bind := n.NetAddrOf(n.host)
	n.server, _ = api.StartService(bind, n, n.log.Infof, true)
}
