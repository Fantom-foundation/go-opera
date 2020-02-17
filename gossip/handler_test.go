package gossip

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/app"
	"github.com/Fantom-foundation/go-lachesis/eventcheck"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
	"github.com/Fantom-foundation/go-lachesis/logger"
	"github.com/Fantom-foundation/go-lachesis/poset"
)

// Tests that events can be retrieved from a remote graph based on user queries.
func TestGetEvents62(t *testing.T) {
	logger.SetTestMode(t)
	testGetEvents(t, lachesis62)
}

func testGetEvents(t *testing.T, protocol int) {
	assertar := assert.New(t)

	var firstEvent *inter.Event
	var someEvent *inter.Event
	var lastEvent *inter.Event
	notExistingEvent := hash.HexToEventHash("0x6099dac580ff18a7055f5c92c2e0717dd4bf9907565df7a8502d0c3dd513b30c")
	pm, _ := newTestProtocolManagerMust(t, 5, 5, nil, func(e *inter.Event) {
		if firstEvent == nil {
			firstEvent = e
		}
		if e.Seq == 3 {
			someEvent = e
		}
		lastEvent = e
	})

	peer, _ := newTestPeer("peer", protocol, pm, true)
	defer peer.close()

	// Create a "random" unknown hash for testing
	var unknown common.Hash
	for i := range unknown {
		unknown[i] = byte(i)
	}
	// Create a batch of tests for various scenarios
	tests := []struct {
		query  []hash.Event   // The query to execute for events retrieval
		expect []*inter.Event // The expected events
	}{
		// A single random event should be retrievable by hash and number too
		{
			[]hash.Event{someEvent.Hash()},
			[]*inter.Event{someEvent},
		},
		// Multiple events should be retrievable in both directions
		{
			[]hash.Event{firstEvent.Hash(), someEvent.Hash(), lastEvent.Hash()},
			[]*inter.Event{firstEvent, someEvent, lastEvent},
		}, {
			[]hash.Event{lastEvent.Hash(), someEvent.Hash(), firstEvent.Hash()},
			[]*inter.Event{lastEvent, someEvent, firstEvent},
		},
		// Check repeated requests
		{
			[]hash.Event{someEvent.Hash(), someEvent.Hash()},
			[]*inter.Event{someEvent, someEvent},
		},
		// Check that non existing events aren't returned
		{
			[]hash.Event{notExistingEvent, someEvent.Hash(), notExistingEvent},
			[]*inter.Event{someEvent},
		},
	}
	// Run each of the tests and verify the results against the chain
	for i, tt := range tests {
		// Send the hash request and verify the response
		if !assertar.NoError(p2p.Send(peer.app, GetEventsMsg, tt.query)) {
			return
		}
		if err := p2p.ExpectMsg(peer.app, EventsMsg, tt.expect); err != nil {
			t.Errorf("test %d: events mismatch: %v", i, err)
		}
		if t.Failed() {
			return
		}
	}
}

func TestBroadcastEvent(t *testing.T) {
	logger.SetTestMode(t)

	var tests = []struct {
		totalPeers                int
		forcedAggressiveBroadcast bool
	}{
		{1, true},
		{1, false},
		{2, true},
		{3, true},
		{4, true},
		{5, false},
		{9, true},
		{12, false},
		{16, false},
		{26, true},
		{100, false},
	}
	for _, test := range tests {
		testBroadcastEvent(t, test.totalPeers, test.forcedAggressiveBroadcast)
	}
}

func testBroadcastEvent(t *testing.T, totalPeers int, forcedAggressiveBroadcast bool) {
	assertar := assert.New(t)

	net := lachesis.FakeNetConfig(genesis.FakeValidators(1, big.NewInt(0), pos.StakeToBalance(1)))
	config := DefaultConfig(net)
	if forcedAggressiveBroadcast {
		config.Protocol.LatencyImportance = 1
		config.Protocol.ThroughputImportance = 0
	} else {
		config.Protocol.LatencyImportance = 0
		config.Protocol.ThroughputImportance = 1
	}
	config.Emitter.EmitIntervals.Min = time.Duration(0)
	config.Emitter.EmitIntervals.Max = time.Duration(0)
	config.Emitter.EmitIntervals.SelfForkProtection = 0
	config.TxPool.Journal = ""

	// create stores
	app := app.NewMemStore()
	state, _, err := app.ApplyGenesis(&net, nil)
	if !assertar.NoError(err) {
		return
	}
	store := NewMemStore()
	genesisAtropos, genesisEvmState, _, err := store.ApplyGenesis(&net, state)
	if !assertar.NoError(err) {
		return
	}
	engineStore := poset.NewMemStore()
	err = engineStore.ApplyGenesis(&net.Genesis, genesisAtropos, genesisEvmState)
	if !assertar.NoError(err) {
		return
	}

	// create consensus engine
	engine := poset.New(net.Dag, engineStore, store)
	engine.Bootstrap(inter.ConsensusCallbacks{})

	// create service
	creator := net.Genesis.Alloc.Validators.Addresses()[0]
	ctx := &node.ServiceContext{
		AccountManager: mockAccountManager(net.Genesis.Alloc.Accounts, creator),
	}
	svc, err := NewService(ctx, &config, store, engine, app)
	assertar.NoError(err)

	// start PM
	pm := svc.pm
	pm.Start(1000)
	pm.synced = 1
	pm.downloader.Terminate() // disable downloader so test would be deterministic
	defer pm.Stop()

	// create peers
	var peers []*testPeer
	for i := 0; i < totalPeers; i++ {
		peer, _ := newTestPeer(fmt.Sprintf("peer %d", i), lachesis62, pm, true)
		defer peer.close()
		peers = append(peers, peer)
	}
	for pm.peers.Len() < totalPeers { // wait until all the peers are registered
		time.Sleep(10 * time.Millisecond)
	}

	// start emitter
	svc.emitter = svc.makeEmitter()
	svc.emitter.SetValidator(creator)

	emittedEvents := make([]*inter.Event, 0)
	for i := 0; i < totalPeers; i++ {
		emitted := svc.emitter.EmitEvent()
		assertar.NotNil(emitted)
		emittedEvents = append(emittedEvents, emitted)
		// check it's broadcasted just after emitting
		for _, peer := range peers {
			if forcedAggressiveBroadcast {
				// aggressive
				assertar.NoError(p2p.ExpectMsg(peer.app, EventsMsg, []*inter.Event{emitted}))
			} else {
				// announce
				assertar.NoError(p2p.ExpectMsg(peer.app, NewEventHashesMsg, []hash.Event{emitted.Hash()}))
			}
			if t.Failed() {
				return
			}
		}
		// broadcast doesn't send to peers who are known to know this event
		assertar.Equal(0, pm.BroadcastEvent(emitted, 0))
	}

	// fresh new peer
	newPeer, _ := newTestPeer(fmt.Sprintf("peer %d", totalPeers), lachesis62, pm, true)
	defer newPeer.close()
	for pm.peers.Len() < totalPeers+1 { // wait until the new peer is registered
		time.Sleep(10 * time.Millisecond)
	}

	// create new event, but send it from new peer
	{
		emitted := svc.emitter.createEvent(nil)
		assertar.NotNil(emitted)
		assertar.NoError(p2p.Send(newPeer.app, NewEventHashesMsg, []hash.Event{emitted.Hash()})) // announce
		// now PM should request it
		assertar.NoError(p2p.ExpectMsg(newPeer.app, GetEventsMsg, []hash.Event{emitted.Hash()})) // request
		if t.Failed() {
			return
		}
		// send it to PM
		assertar.NoError(p2p.Send(newPeer.app, EventsMsg, []*inter.Event{emitted}))
		// PM should broadcast it to all other peer except newPeer
		for _, peer := range peers {
			if forcedAggressiveBroadcast {
				// aggressive
				assertar.NoError(p2p.ExpectMsg(peer.app, EventsMsg, []*inter.Event{emitted}))
			} else {
				// announce
				assertar.NoError(p2p.ExpectMsg(peer.app, NewEventHashesMsg, []hash.Event{emitted.Hash()}))
			}
			if t.Failed() {
				return
			}
			assertar.True(svc.store.HasEvent(emitted.Hash()), emitted.Hash().String())
		}
		emittedEvents = append(emittedEvents, emitted)
	}

	// peers request the event. check it at the end, so we known that nothing was sent before
	for _, emitted := range emittedEvents {
		for _, peer := range append(peers, newPeer) {
			assertar.NoError(p2p.Send(peer.app, GetEventsMsg, []hash.Event{emitted.Hash()})) // request
			assertar.NoError(p2p.ExpectMsg(peer.app, EventsMsg, []*inter.Event{emitted}))    // response
			if t.Failed() {
				return
			}
		}
	}
}

func mockAccountManager(accs genesis.Accounts, unlock ...common.Address) *accounts.Manager {
	return accounts.NewManager(
		&accounts.Config{InsecureUnlockAllowed: true},
		genesis.NewAccountsBackend(accs, unlock...),
	)
}

func mockCheckers(epoch idx.Epoch, net *lachesis.Config, engine Consensus, s *Store, a *app.Store) *eventcheck.Checkers {
	heavyCheckReader := &HeavyCheckReader{}
	heavyCheckReader.Addrs.Store(ReadEpochPubKeys(a, epoch))
	gasPowerCheckReader := &GasPowerCheckReader{}
	gasPowerCheckReader.Ctx.Store(ReadGasPowerContext(s, a, engine.GetValidators(), engine.GetEpoch(), &net.Economy))
	return makeCheckers(net, heavyCheckReader, gasPowerCheckReader, engine, s)
}
