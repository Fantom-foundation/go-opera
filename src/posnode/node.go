package posnode

import (
	"crypto/ecdsa"

	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/network"
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
	downloads
	discovery
	logger
}

// New creates node.
// It does not start any process.
func New(host string, key *ecdsa.PrivateKey, s *Store, c Consensus, conf *Config, listen network.ListenFunc, opts ...grpc.DialOption) *Node {
	if key == nil {
		var err error
		key, err = crypto.GenerateECDSAKey()
		if err != nil {
			panic(err)
		}
	}

	if conf == nil {
		conf = DefaultConfig()
	}

	n := Node{
		ID:        CalcPeerID(&key.PublicKey),
		key:       key,
		pub:       &key.PublicKey,
		store:     s,
		consensus: c,
		host:      host,
		conf:      *conf,
		service:   service{listen, nil},
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
	n.StartEventEmission()
}

// Stop stops all node services.
func (n *Node) Stop() {
	n.StopEventEmission()
	n.StopGossip()
	n.StopDiscovery()
	n.StopService()
}

// PubKey returns public key.
func (n *Node) PubKey() *ecdsa.PublicKey {
	pk := *n.pub
	return &pk
}

// Host returns host.
func (n *Node) Host() string {
	return n.host
}

// AsPeer returns nodes peer info.
func (n *Node) AsPeer() *Peer {
	return &Peer{
		ID:     n.ID,
		PubKey: n.pub,
		Host:   n.host,
	}
}

/*
 * Utils:
 */

// CalcPeerID returns peer id from pub key.
func CalcPeerID(pub *ecdsa.PublicKey) hash.Peer {
	return CalcPeerInfoID(common.FromECDSAPub(pub))
}

// CalcPeerInfoID returns peer id from pub key bytes.
func CalcPeerInfoID(pub []byte) hash.Peer {
	return hash.Peer(hash.Of(pub))
}

// FakeClient returns dialer for fake network.
func FakeClient(host string) grpc.DialOption {
	dialer := network.FakeDialer(host)
	return grpc.WithContextDialer(dialer)
}
