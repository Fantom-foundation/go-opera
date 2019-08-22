package gossip

import (
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
)

// Tests that events can be retrieved from a remote graph based on user queries.
func TestGetEvents62(t *testing.T) { testGetEvents(t, fantom62) }

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
	var tests = []struct {
		totalPeers        int
		broadcastExpected int
		allowAggressive   bool
	}{
		{1, 1, true},
		{1, 1, false},
		{2, 2, true},
		{3, 3, true},
		{4, 4, true},
		{5, 4, false},
		{9, 4, false},
		{12, 4, false},
		{16, 4, false},
		{26, 5, false},
		{100, 10, false},
	}
	for _, test := range tests {
		testBroadcastEvent(t, test.totalPeers, test.broadcastExpected, test.allowAggressive)
	}
}

func testBroadcastEvent(t *testing.T, totalPeers, broadcastExpected int, allowAggressive bool) {
	if allowAggressive && totalPeers > minBroadcastPeers {
		t.Error("Wrong testcase: allowAggressive must be false if totalPeers > minBroadcastPeers (because we'll broadcast only to a subset)")
	}

	assertar := assert.New(t)

	network, nodes, keys := lachesis.FakeNetConfig(1)
	config := DefaultConfig(network)
	config.Emitter.MinEmitInterval = 10 * time.Millisecond
	config.Emitter.MaxEmitInterval = 10 * time.Millisecond
	config.ForcedBroadcast = allowAggressive

	var (
		evmux = new(event.TypeMux)
		store = NewMemStore()
	)

	privateKey := keys[0]
	me := nodes[0]

	engineStore := poset.NewMemStore()
	assertar.NoError(engineStore.ApplyGenesis(&network.Genesis))

	engine := poset.New(engineStore, store)
	engine.Bootstrap(nil)

	svc, err := NewService(config, evmux, store, engine)
	assertar.NoError(err)

	// start PM
	pm := svc.pm
	pm.Start(1000)
	pm.synced = 1
	defer pm.Stop()

	// create peers
	var peers []*testPeer
	for i := 0; i < totalPeers; i++ {
		peer, _ := newTestPeer(fmt.Sprintf("peer %d", i), fantom62, pm, true)
		defer peer.close()
		peers = append(peers, peer)
	}
	for pm.peers.Len() < totalPeers { // wait until all the peers are registered
		time.Sleep(10 * time.Millisecond)
	}

	// start emitter
	svc.privateKey = privateKey
	svc.me = me
	svc.emitter = svc.makeEmitter()

	emittedEvents := make([]*inter.Event, 0)
	for i := 0; i < broadcastExpected; i++ {
		emitted := svc.emitter.EmitEvent()
		emittedEvents = append(emittedEvents, emitted)
		// check it's broadcasted just after emitting
		for _, peer := range peers {
			if allowAggressive {
				assertar.NoError(p2p.ExpectMsg(peer.app, EventsMsg, []*inter.Event{emitted})) // aggressive
			} else {
				assertar.NoError(p2p.ExpectMsg(peer.app, NewEventHashesMsg, []hash.Event{emitted.Hash()})) // announce
			}
			if t.Failed() {
				return
			}
		}
		// broadcast doesn't send to peers who are known to know this event
		assertar.Equal(0, pm.BroadcastEvent(emitted, false))
	}

	// fresh new peer
	newPeer, _ := newTestPeer(fmt.Sprintf("peer %d", totalPeers), fantom62, pm, true)
	defer newPeer.close()
	for pm.peers.Len() < totalPeers+1 { // wait until the new peer is registered
		time.Sleep(10 * time.Millisecond)
	}

	// create new event, but send it from new peer
	{
		emitted := svc.emitter.createEvent()
		assertar.NoError(p2p.Send(newPeer.app, NewEventHashesMsg, []hash.Event{emitted.Hash()})) // announce
		// now PM should request it
		assertar.NoError(p2p.ExpectMsg(newPeer.app, GetEventsMsg, []hash.Event{emitted.Hash()})) // request
		if t.Failed() {
			return
		}
		// send it to PM
		assertar.NoError(p2p.Send(newPeer.app, EventsMsg, []*inter.Event{emitted}))
		// PM should broadcast it to all other peer except newPeer, non-aggressively
		for _, peer := range peers {
			assertar.NoError(p2p.ExpectMsg(peer.app, NewEventHashesMsg, []hash.Event{emitted.Hash()}))
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
