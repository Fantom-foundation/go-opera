package posnode

import (
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/network"
)

// Node is a Lachesis node implementation.
type Node struct {
	ID        hash.Peer
	key       *common.PrivateKey
	pub       *common.PublicKey
	store     *Store
	consensus Consensus
	host      string
	conf      Config

	service
	connPool
	peers
	parents
	emitter
	gossip
	downloads
	discovery
	builtin
	logger
}

// New creates node.
// It does not start any process.
func New(host string, key *common.PrivateKey, s *Store, c Consensus, conf *Config, listen network.ListenFunc, opts ...grpc.DialOption) *Node {
	if key == nil {
		key = crypto.GenerateKey()
	}

	if conf == nil {
		conf = DefaultConfig()
	}

	n := Node{
		ID:        hash.PeerOfPubkey(key.Public()),
		key:       key,
		pub:       key.Public(),
		store:     s,
		consensus: c,
		host:      host,
		conf:      *conf,
		service:   service{listen, nil},
		connPool:  connPool{opts: opts},
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
func (n *Node) PubKey() *common.PublicKey {
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

// LastEventOf returns last event of peer.
func (n *Node) LastEventOf(peer hash.Peer) *inter.Event {
	i := n.store.GetPeerHeight(peer)
	if i == 0 {
		return nil
	}

	return n.EventOf(peer, i)
}

// EventOf returns i-th event of peer.
func (n *Node) EventOf(peer hash.Peer, i uint64) *inter.Event {
	h := n.store.GetEventHash(peer, i)
	if h == nil {
		n.log.Errorf("no event hash for (%s,%d) in store", peer.String(), i)
		return nil
	}

	e := n.store.GetEvent(*h)
	if e == nil {
		n.log.Errorf("no event %s in store", e.String())
	}

	return e
}

// GetID returns node id.
func (n *Node) GetID() hash.Peer {
	return n.ID
}

/*
 * Utils:
 */

// FakeClient returns dialer for fake network.
func FakeClient(host string) grpc.DialOption {
	dialer := network.FakeDialer(host)
	return grpc.WithContextDialer(dialer)
}
