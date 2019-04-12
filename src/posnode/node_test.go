package posnode

//NOTE: mockgen does not work properly out of GOPATH
//go:generate mockgen -package=posnode -source=consensus.go -destination=mock_test.go Consensus

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/Fantom-foundation/go-lachesis/src/network"
)

func ExampleNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consensus := NewMockConsensus(ctrl)

	store := NewMemStore()

	n := NewForTests("server.fake", store, consensus)

	n.Start()
	defer n.Stop()

	select {}
}

// NewForTests creates node with fake network client.
func NewForTests(host string, s *Store, c Consensus) *Node {
	return New(host, nil, s, c, nil, network.FakeListener, FakeClient(host))
}
