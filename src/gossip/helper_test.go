package gossip

import (
	"crypto/ecdsa"
	"crypto/rand"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"math/big"
	"sort"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

var (
	testBankKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	testBank       = crypto.PubkeyToAddress(testBankKey.PublicKey)
)

// newTestProtocolManager creates a new protocol manager for testing purposes,
// with the given number of events already known from each node
func newTestProtocolManager(nodesNum int, eventsNum int, newtx chan<- []*types.Transaction, onNewEvent func(e *inter.Event)) (*ProtocolManager, *Store, error) {
	var (
		evmux = new(event.TypeMux)
		store = NewMemStore()
	)

	nodes := inter.GenNodes(nodesNum)
	balances := make(map[hash.Peer]inter.Stake, len(nodes))
	for _, addr := range nodes {
		balances[addr] = inter.Stake(1)
	}

	engineStore := posposet.NewMemStore()
	err := engineStore.ApplyGenesis(balances, 0)
	if err != nil {
		return nil, nil, err
	}

	engine := posposet.New(engineStore, store)
	engine.Bootstrap()

	config := &Config{}
	config.NetworkId = 1

	pm, err := NewProtocolManager(config.Dag, downloader.FullSync, config.NetworkId, evmux, &testTxPool{added: newtx}, new(sync.RWMutex), store, engine)
	if err != nil {
		return nil, nil, err
	}

	buildEvent := func(e *inter.Event) *inter.Event {
		e.Epoch = 1
		return engine.Prepare(e)
	}
	_onNewEvent := func(e *inter.Event) {
		store.SetEvent(e)
		err = engine.ProcessEvent(e)
		if err != nil {
			panic(err)
		}
		if onNewEvent != nil {
			onNewEvent(e)
		}
	}

	inter.GenEventsByNode(nodes, eventsNum, 3, buildEvent, _onNewEvent, nil)

	pm.Start(1000)
	return pm, store, nil
}

// newTestProtocolManagerMust creates a new protocol manager for testing purposes,
// with the given number of events already known from each peer. In case of an error, the constructor force-
// fails the test.
func newTestProtocolManagerMust(t *testing.T, nodes int, events int, newtx chan<- []*types.Transaction, onNewEvent func(e *inter.Event)) (*ProtocolManager, *Store) {
	pm, db, err := newTestProtocolManager(nodes, events, newtx, onNewEvent)
	if err != nil {
		t.Fatalf("Failed to create protocol manager: %v", err)
	}
	return pm, db
}

// testTxPool is a fake, helper transaction pool for testing purposes
type testTxPool struct {
	txFeed event.Feed
	pool   []*types.Transaction        // Collection of all transactions
	added  chan<- []*types.Transaction // Notification channel for new transactions

	lock sync.RWMutex // Protects the transaction pool
}

// AddRemotes appends a batch of transactions to the pool, and notifies any
// listeners if the addition channel is non nil
func (p *testTxPool) AddRemotes(txs []*types.Transaction) []error {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.pool = append(p.pool, txs...)
	if p.added != nil {
		p.added <- txs
	}
	return make([]error, len(txs))
}

// Pending returns all the transactions known to the pool
func (p *testTxPool) Pending() (map[common.Address]types.Transactions, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	batches := make(map[common.Address]types.Transactions)
	for _, tx := range p.pool {
		from, _ := types.Sender(types.HomesteadSigner{}, tx)
		batches[from] = append(batches[from], tx)
	}
	for _, batch := range batches {
		sort.Sort(types.TxByNonce(batch))
	}
	return batches, nil
}

func (p *testTxPool) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return p.txFeed.Subscribe(ch)
}

// newTestTransaction create a new dummy transaction.
func newTestTransaction(from *ecdsa.PrivateKey, nonce uint64, datasize int) *types.Transaction {
	tx := types.NewTransaction(nonce, common.Address{}, big.NewInt(0), 100000, big.NewInt(0), make([]byte, datasize))
	tx, _ = types.SignTx(tx, types.HomesteadSigner{}, from)
	return tx
}

// testPeer is a simulated peer to allow testing direct network calls.
type testPeer struct {
	net p2p.MsgReadWriter // Network layer reader/writer to simulate remote messaging
	app *p2p.MsgPipeRW    // Application layer reader/writer to simulate the local side
	*peer
}

// newTestPeer creates a new peer registered at the given protocol manager.
func newTestPeer(name string, version int, pm *ProtocolManager, shake bool) (*testPeer, <-chan error) {
	// Create a message pipe to communicate through
	app, net := p2p.MsgPipe()

	// Generate a random id and create the peer
	var id enode.ID
	rand.Read(id[:])

	peer := pm.newPeer(version, p2p.NewPeer(id, name, nil), net)

	// Start the peer on a new thread
	errc := make(chan error, 1)
	go func() {
		select {
		case pm.newPeerCh <- peer:
			errc <- pm.handle(peer)
		case <-pm.quitSync:
			errc <- p2p.DiscQuitting
		}
	}()
	tp := &testPeer{app: app, net: net, peer: peer}
	// Execute any implicitly requested handshakes and return
	if shake {
		var (
			genesis  = pm.engine.GetGenesisHash()
			progress = PeerProgress{
				Epoch: pm.engine.CurrentSuperFrameN(),
			}
		)
		tp.handshake(nil, progress, genesis)
	}
	return tp, errc
}

// handshake simulates a trivial handshake that expects the same state from the
// remote side as we are simulating locally.
func (p *testPeer) handshake(t *testing.T, progress PeerProgress, genesis hash.Hash) {
	msg := &statusData{
		ProtocolVersion: uint32(p.version),
		NetworkId:       1,
		Progress:        progress,
		Genesis:         genesis,
	}
	if err := p2p.ExpectMsg(p.app, StatusMsg, msg); err != nil {
		t.Fatalf("status recv: %v", err)
	}
	if err := p2p.Send(p.app, StatusMsg, msg); err != nil {
		t.Fatalf("status send: %v", err)
	}
}

// close terminates the local side of the peer, notifying the remote protocol
// manager of termination.
func (p *testPeer) close() {
	p.app.Close()
}
