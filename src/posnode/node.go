package posnode

import (
	"crypto/ecdsa"
	"time"

	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

const (
	emitterTickInterval = time.Minute
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
	emitter
	gossip
	discovery
	logger
}

// New creates node.
// It does not start any process.
func New(host string, key *ecdsa.PrivateKey, s *Store, c Consensus, conf *Config, opts ...grpc.DialOption) *Node {
	n := Node{
		ID:        CalcPeerID(&key.PublicKey),
		key:       key,
		pub:       &key.PublicKey,
		store:     s,
		consensus: c,
		host:      host,
		conf:      *conf,
		client:    client{opts},
		logger:    newLogger(host),
	}

	return &n
}

// Start starts all node services.
func (n *Node) Start() {
	n.StartService()
	n.StartDiscovery()
	n.StartGossip(n.conf.GossipThreads)
	n.StartEmit()
}

// Stop stops all node services.
func (n *Node) Stop() {
	n.StopGossip()
	n.StopDiscovery()
	n.StopService()
	n.StopEmit()
}

// CalcPeerID returns peer id from pub key.
func CalcPeerID(pub *ecdsa.PublicKey) hash.Peer {
	return CalcPeerInfoID(common.FromECDSAPub(pub))
}

// CalcPeerInfoID returns peer id from pub key bytes.
func CalcPeerInfoID(pub []byte) hash.Peer {
	return hash.Peer(hash.Of(pub))
}

// ToPeer returns hash.Peer
func (n *Node) ToPeer() hash.Peer {
	return hash.Peer(hash.Of(common.FromECDSAPub(n.pub)))
}
