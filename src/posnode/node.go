package posnode

import (
	"crypto/ecdsa"

	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

// Node is a Lachesis node implementation.
type Node struct {
	ID  common.Hash
	key *ecdsa.PrivateKey
	pub *ecdsa.PublicKey

	consensus Consensus

	server *grpc.Server
	dialer Dialer
}

// New creates node.
func New(key *ecdsa.PrivateKey, c Consensus, dialer Dialer) *Node {
	return &Node{
		ID:  common.BytesToHash(crypto.FromECDSAPub(&key.PublicKey)),
		key: key,
		pub: &key.PublicKey,

		consensus: c,
		dialer:    dialer,
	}
}

// Shutdown stops node.
func (n *Node) Shutdown() {

}
