package posnode

import (
	"github.com/Fantom-foundation/go-lachesis/src/cryptoaddr"
	"sync"

	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/ordering"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/network"
)

// Node is a Lachesis node implementation.
type Node struct {
	ID        hash.Peer
	key       *crypto.PrivateKey
	store     *Store
	consensus Consensus
	host      string
	conf      Config

	onNewEvent func(*inter.Event)

	service
	connPool
	superFrame
	peers
	parents
	emitter
	gossip
	downloads
	discovery
	builtin

	logger.Instance
}

// New creates node.
// It does not start any process.
func New(host string, key *crypto.PrivateKey, s *Store, c Consensus, conf *Config, listen network.ListenFunc, opts ...grpc.DialOption) *Node {
	if key == nil {
		nKey, err := crypto.GenerateKey()
		if err != nil {
			panic(err)
		}
		key = nKey
	}

	if conf == nil {
		conf = DefaultConfig()
	}

	n := Node{
		ID:        cryptoaddr.AddressOf(key.Public()),
		key:       key,
		store:     s,
		consensus: c,
		host:      host,
		conf:      *conf,

		service:  service{"", listen, nil},
		connPool: connPool{opts: opts},

		Instance: logger.MakeInstance(),
	}

	orderThenSave := ordering.EventBuffer(ordering.Callback{

		Process: n.saveNewEvent,

		Drop: func(e *inter.Event, err error) {
			n.Warn(err.Error() + ", so rejected")
		},

		Exists: func(h hash.Event) *inter.Event {
			return n.store.GetEvent(h)
		},
	})

	var save sync.Mutex
	n.onNewEvent = func(e *inter.Event) {
		// TODO: replace mutex with chan
		save.Lock()
		defer save.Unlock()
		orderThenSave(e)
	}

	return &n
}

// saveNewEvent writes event to store, indexes and consensus.
// It is not safe for concurrent use.
func (n *Node) saveNewEvent(e *inter.Event) {
	n.Debugf("save new event")

	n.store.SetEvent(e)
	n.store.SetEventHash(e.Creator, e.SfNum, e.Seq, e.Hash())
	// NOTE: doubled txns from evil event could override existing index!
	// TODO: decision
	n.store.SetTxnsEvent(e.Hash(), e.Creator, e.InternalTransactions...)

	n.store.SetPeerHeight(e.Creator, e.SfNum, e.Seq)
	n.setLast(e)
	n.pushPotentialParent(e)

	if n.consensus != nil {
		n.consensus.PushEvent(e.Hash())
	}

	countTotalEvents.Inc(1)
}

// GetInternalTxn finds transaction ant its event if exists.
func (n *Node) GetInternalTxn(idx hash.Transaction) (
	txn *inter.InternalTransaction,
	event *inter.Event,
) {
	txn = n.internalTxnMempool(idx)
	if txn != nil {
		return
	}

	event = n.store.GetTxnsEvent(idx)
	if event != nil {
		txn = event.FindInternalTxn(idx)
	}

	return
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

	n.stopClient()
}

// Host returns host.
func (n *Node) Host() string {
	return n.host
}

// AsPeer returns nodes peer info.
func (n *Node) AsPeer() *Peer {
	return &Peer{
		ID:     n.ID,
		Host:   n.host,
	}
}

// LastEventOf returns last event of peer.
func (n *Node) LastEventOf(peer hash.Peer, sf idx.SuperFrame) *inter.Event {
	i := n.store.GetPeerHeight(peer, sf)
	if i == 0 {
		return nil
	}

	return n.EventOf(peer, sf, i)
}

// EventOf returns i-th event of peer.
func (n *Node) EventOf(peer hash.Peer, sf idx.SuperFrame, i idx.Event) *inter.Event {
	h := n.store.GetEventHash(peer, sf, i)
	if h == nil {
		n.Errorf("no event hash for (%s,%d-%d) in store", peer.String(), sf, i)
		return nil
	}

	e := n.store.GetEvent(*h)
	if e == nil {
		n.Errorf("no event in store of %d", i)
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
