package posnode

import (
	"crypto/ecdsa"
	"fmt"
	"sync"

	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/api"
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

// SetDialOpts allows to reset grpc client dial options.
func SetDialOpts(opts ...grpc.DialOption) Option {
	return func(n *Node) {
		n.client.opts = opts
	}
}

// SetDiscoveryMem sets discovery storage as memory.
func SetDiscoveryMem(n *Node) {
	n.discovery.store = &memDiscovery{
		RWMutex:     sync.RWMutex{},
		discoveries: make(map[string]*Discovery),
	}
}

// SetDiscoveryBadger sets discovery storage as badger.
func SetDiscoveryBadger(n *Node) {
	n.discovery.store = n.store
}

// Option allows to change Node behaviour.
type Option func(*Node)

// New creates node.
func New(host string, key *ecdsa.PrivateKey, s *Store, c Consensus, conf *Config, opts ...Option) *Node {
	n := Node{
		ID:        CalcNodeID(&key.PublicKey),
		key:       key,
		pub:       &key.PublicKey,
		store:     s,
		consensus: c,
		host:      host,
		conf:      *conf,

		discovery: discovery{
			store: s,
		},

		peers:  initPeers(s),
		logger: newLogger(host),
	}

	for _, opt := range opts {
		opt(&n)
	}

	// Add self into peers store
	peerInfo := api.PeerInfo{
		ID:     n.ID.Hex(),
		PubKey: common.FromECDSAPub(n.pub),
		Host:   fmt.Sprintf("%s:%d", host, conf.Port),
	}
	n.store.SetPeer(WireToPeer(&peerInfo))

	return &n
}

// Shutdown stops node.
func (n *Node) Shutdown() {
	n.log.Info("shutdown")
}

// CalcNodeID returns peer from pub key.
func CalcNodeID(pub *ecdsa.PublicKey) hash.Peer {
	return hash.Peer(hash.Of(common.FromECDSAPub(pub)))
}
