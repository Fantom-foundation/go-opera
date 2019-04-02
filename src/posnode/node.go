package posnode

import (
	"crypto/ecdsa"

	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// Node is a Lachesis node implementation.
type Node struct {
	ID        hash.Peer
	key       *ecdsa.PrivateKey
	pub       *ecdsa.PublicKey
	store     *Store
	consensus Consensus
	host      string
	conf      Config

	service
	client
	peers
	gossip
	discovery
	logger
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
		peers:  initPeers(s),
		logger: newLogger(host),
	}
}

// Shutdown stops node.
func (n *Node) Shutdown() {
	n.log.Info("shutdown")
}

/*
* Utils:
 */

func CalcNodeID(pub *ecdsa.PublicKey) hash.Peer {
	return hash.Peer(hash.Of(common.FromECDSAPub(pub)))
}
