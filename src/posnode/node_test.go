package posnode

//NOTE: mockgen does not work properly out of GOPATH
//go:generate mockgen -package=posnode -source=consensus.go -destination=mock_test.go Consensus

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/network"
)

func TestNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consensus := NewMockConsensus(ctrl)

	store := NewMemStore()

	n := NewForTests("node001", store, consensus)
	defer n.Shutdown()

	n.StartServiceForTests()
	defer n.StopService()

	n.StartDiscovery()
	defer n.StopDiscovery()

	n.StartGossip(4)
	defer n.StopGossip()
	<-time.After(5 * time.Second)
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
	bind := net.JoinHostPort(n.host, strconv.Itoa(n.conf.Port))
	listener := network.FakeListener(bind)
	n.startService(listener)
}
