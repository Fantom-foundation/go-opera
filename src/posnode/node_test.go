package posnode

//NOTE: mockgen does not work properly out of GOPATH
//go:generate mockgen -package=posnode -source=consensus.go -destination=mock_test.go Consensus

import (
	"github.com/Fantom-foundation/go-lachesis/src/network"
)

// NewForTests creates node with fake network client.
func NewForTests(host string, s *Store, c Consensus) *Node {
	n := New(host, nil, s, c, nil, network.FakeListener, FakeClient(host))
	if s != nil {
		s.SetName(host)
	}
	return n
}
