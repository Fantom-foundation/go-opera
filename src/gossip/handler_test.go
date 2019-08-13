package gossip

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
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
