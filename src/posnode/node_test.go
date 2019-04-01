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

func ExampleNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consensus := NewMockConsensus(ctrl)

	store := NewMemStore()

	n := NewForTests("server.fake", store, consensus)
	defer n.Shutdown()

	peerInfoAsksChan := make(chan struct{})
	peerInfoAsked = func() {
		peerInfoAsksChan <- struct{}{}
	}

	n.StartServiceForTests()
	defer n.StopService()

	n.StartDiscovery()
	defer n.StopDiscovery()

	n.StartGossip(4)
	defer n.StopGossip()

	// should make 1 fail call for n.AskPeerInfo
	select {
	case <-peerInfoAsksChan:
	case <-time.After(time.Second):
		t.Error("should ask peer info")
	}
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

	return New(host, key, s, c, DefaultConfig(), SetDialOpts(opts), SetDiscoveryMem)
}

// StartServiceForTests starts node service.
// It should be called once.
func (n *Node) StartServiceForTests() {
	bind := net.JoinHostPort(n.host, strconv.Itoa(n.conf.Port))
	listener := network.FakeListener(bind)
	n.startService(listener)
}
