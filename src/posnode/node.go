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
	queue
	parentQueue
}

// New creates node.
func New(host string, key *ecdsa.PrivateKey, s *Store, c Consensus, conf *Config, opts ...grpc.DialOption) *Node {
	n := Node{
		ID:          CalcPeerID(&key.PublicKey),
		key:         key,
		pub:         &key.PublicKey,
		store:       s,
		consensus:   c,
		host:        host,
		conf:        *conf,
		client:      client{opts},
		logger:      newLogger(host),
		queue:       initQueue(),
		parentQueue: initParentQueue(),
	}

	return &n
}

// Shutdown stops node.
func (n *Node) Shutdown() {
	n.log.Info("shutdown")
}

// CalcPeerID returns peer id from pub key.
func CalcPeerID(pub *ecdsa.PublicKey) hash.Peer {
	return CalcPeerInfoID(common.FromECDSAPub(pub))
}

// CalcPeerInfoID returns peer id from pub key bytes.
func CalcPeerInfoID(pub []byte) hash.Peer {
	return hash.Peer(hash.Of(pub))
}
