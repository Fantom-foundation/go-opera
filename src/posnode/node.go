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
	queue
	parentQueue
}

// New creates node.
func New(host string, key *ecdsa.PrivateKey, s *Store, c Consensus, conf *Config, opts ...grpc.DialOption) *Node {
	n := Node{
		ID:        CalcNodeID(&key.PublicKey),
		key:       key,
		pub:       &key.PublicKey,
		store:     s,
		consensus: c,
		host:      host,
		conf:      *conf,
		client:    client{opts},
		discovery: discovery{
			store: &discoveries{
				RWMutex:     sync.RWMutex{},
				discoveries: make(map[string]*Discovery),
			},
		},

		peers:       initPeers(s),
		logger:      newLogger(host),
		queue:       initQueue(),
		parentQueue: initParentQueue(),
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
