package posnode

import (
	"github.com/Fantom-foundation/go-lachesis/src/network"
)

// NewForTests creates node with fake network client.
func NewForTests(host string, s *Store, c Consensus) *Node {
	if s == nil {
		s = NewMemStore()
	}

	n := New(host, nil, s, c, nil, network.FakeListener, FakeClient(host))
	if s != nil {
		s.SetName(host)
	}
	return n
}
