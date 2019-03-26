package posnode

import (
	"crypto/ecdsa"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

// Node is a Lachesis node implementation.
type Node struct {
	ID    common.Address
	key   *ecdsa.PrivateKey
	pub   *ecdsa.PublicKey
	store *Store

	consensus Consensus

	service
	client
	gossip
	discovery

	connectedPeers map[common.Address]bool

	knownHeights map[string]uint64 // [peerID]lastIndex // TODO: to store?
}

// New creates node.
func New(key *ecdsa.PrivateKey, s *Store, c Consensus, opts ...grpc.DialOption) *Node {
	return &Node{
		ID:    CalcNodeID(&key.PublicKey),
		key:   key,
		pub:   &key.PublicKey,
		store: s,

		consensus: c,

		client: client{opts},
	}
}

// Shutdown stops node.
func (n *Node) Shutdown() {
	n.log().Info("shutdown")
}

/*
* Utils:
 */

func (n *Node) log() *logrus.Entry {
	return GetLogger(n.ID, "")
}

func CalcNodeID(pub *ecdsa.PublicKey) common.Address {
	return common.BytesToAddress(crypto.FromECDSAPub(pub))
}
