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
	ID        common.Address
	key       *ecdsa.PrivateKey
	pub       *ecdsa.PublicKey
	store     *Store
	consensus Consensus
	host      string
	conf      Config

	service
	client
	gossip
	discovery
}

// New creates node.
func New(host string, key *ecdsa.PrivateKey, s *Store, c Consensus, conf *Config, opts ...grpc.DialOption) *Node {
	return &Node{
		ID:        CalcNodeID(&key.PublicKey),
		key:       key,
		pub:       &key.PublicKey,
		store:     s,
		consensus: c,
		host:      host,
		conf:      *conf,

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
	return GetLogger(n.ID, n.host)
}

func CalcNodeID(pub *ecdsa.PublicKey) common.Address {
	return common.BytesToAddress(crypto.FromECDSAPub(pub))
}
