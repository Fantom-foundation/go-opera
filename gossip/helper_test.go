package gossip

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/Fantom-foundation/go-lachesis/app"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
	"github.com/Fantom-foundation/go-lachesis/poset"
)

// newTestProtocolManager creates a new protocol manager for testing purposes,
// with the given number of events already known from each node
func newTestProtocolManager(nodesNum int, eventsNum int, newtx chan<- []*types.Transaction, onNewEvent func(e *inter.Event)) (*ProtocolManager, *Store, error) {
	net := lachesis.FakeNetConfig(genesis.FakeValidators(nodesNum, big.NewInt(0), pos.StakeToBalance(1)))

	config := DefaultConfig(net)
	config.TxPool.Journal = ""

	engineStore := poset.NewMemStore()
	err := engineStore.ApplyGenesis(&net.Genesis, hash.Event{}, common.Hash{})
	if err != nil {
		return nil, nil, err
	}

	app := app.NewMemStore()
	state, _, err := app.ApplyGenesis(&net, nil)
	if err != nil {
		return nil, nil, err
	}

	store := NewMemStore()
	_, _, _, err = store.ApplyGenesis(&net, state)
	if err != nil {
		return nil, nil, err
	}

	engine := poset.New(net.Dag, engineStore, store)
	engine.Bootstrap(inter.ConsensusCallbacks{})

	pm, err := NewProtocolManager(
		&config,
		nil,
		&dummyTxPool{added: newtx},
		new(sync.RWMutex),
		mockCheckers(1, &net, engine, store, app),
		store,
		engine,
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	inter.ForEachRandEvent(net.Genesis.Alloc.Validators.Validators().IDs(), eventsNum, 3, nil, inter.ForEachEvent{
		Process: func(e *inter.Event, name string) {
			store.SetEvent(e)
			err = engine.ProcessEvent(e)
			if err != nil {
				panic(err)
			}
			if onNewEvent != nil {
				onNewEvent(e)
			}
		},
		Build: func(e *inter.Event, name string) *inter.Event {
			e.Epoch = 1
			return engine.Prepare(e)
		},
	})

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
			pm.wg.Add(1)
			defer pm.wg.Done()
			errc <- pm.handle(peer)
		case <-pm.quitSync:
			errc <- p2p.DiscQuitting
		}
	}()
	tp := &testPeer{app: app, net: net, peer: peer}
	// Execute any implicitly requested handshakes and return
	if shake {
		var (
			genesis       = pm.engine.GetGenesisHash()
			blockI, block = pm.engine.LastBlock()
			epoch         = pm.engine.GetEpoch()
			myProgress    = &PeerProgress{
				Epoch:        epoch,
				NumOfBlocks:  blockI,
				LastBlock:    block,
				LastPackInfo: pm.store.GetPackInfoOrDefault(epoch, pm.store.GetPacksNumOrDefault(epoch)-1),
			}
		)
		tp.handshake(nil, myProgress, genesis)
	}
	return tp, errc
}

// handshake simulates a trivial handshake that expects the same state from the
// remote side as we are simulating locally.
func (p *testPeer) handshake(t *testing.T, progress *PeerProgress, genesis common.Hash) {
	msg := &ethStatusData{
		ProtocolVersion:   uint32(p.version),
		NetworkID:         lachesis.FakeNetworkID,
		Genesis:           genesis,
		DummyTD:           big.NewInt(int64(progress.NumOfBlocks)), // for ETH clients
		DummyCurrentBlock: common.Hash(progress.LastBlock),
	}
	if err := p2p.ExpectMsg(p.app, EthStatusMsg, msg); err != nil {
		t.Fatalf("status recv: %v", err)
	}
	if err := p2p.Send(p.app, EthStatusMsg, msg); err != nil {
		t.Fatalf("status send: %v", err)
	}
	if err := p2p.ExpectMsg(p.app, ProgressMsg, progress); err != nil {
		t.Fatalf("progress recv: %v", err)
	}
	if err := p2p.Send(p.app, ProgressMsg, progress); err != nil {
		t.Fatalf("progress send: %v", err)
	}
}

// close terminates the local side of the peer, notifying the remote protocol
// manager of termination.
func (p *testPeer) close() {
	p.app.Close()
}
