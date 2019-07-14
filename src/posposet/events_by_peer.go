package posposet

import (
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// TODO: make EventsByPeer internal

type (
	// EventsByNode is a event hashes grouped by creator.
	// ( creator --> event hashes )
	EventsByPeer map[hash.Peer]hash.Events
)

/*
 * eventsByNode's methods:
 */

// Add unions events into one.
func (ee EventsByPeer) Add(events EventsByPeer) (changed bool) {
	for creator, hashes := range events {
		if ee[creator] == nil {
			ee[creator] = hash.Events{}
		}
		if ee[creator].Add(hashes.Slice()...) {
			changed = true
		}
	}
	return
}

// AddOne appends one event.
func (ee EventsByPeer) AddOne(event hash.Event, creator hash.Peer) (changed bool) {
	if ee[creator] == nil {
		ee[creator] = hash.Events{}
	}
	if ee[creator].Add(event) {
		changed = true
	}
	return
}

// Contains returns true if event of node is in.
func (ee EventsByPeer) Contains(node hash.Peer, event hash.Event) bool {
	return ee[node] != nil && ee[node].Contains(event)
}

// Each returns range of all events.
func (ee EventsByPeer) Each() map[hash.Event]hash.Peer {
	res := make(map[hash.Event]hash.Peer)
	for creator, events := range ee {
		for e := range events {
			res[e] = creator
		}
	}
	return res
}

// String returns human readable string representation.
func (ee EventsByPeer) String() string {
	var ss []string
	for node, roots := range ee {
		ss = append(ss, node.String()+":"+roots.String())
	}
	return "byNode{" + strings.Join(ss, ", ") + "}"
}

// ToWire converts to simple slice.
func (ee EventsByPeer) ToWire() []*wire.EventDescr {
	var arr []*wire.EventDescr
	for creator, hh := range ee {
		for hash_ := range hh {
			arr = append(arr, &wire.EventDescr{
				Creator: creator.Bytes(),
				Hash:    hash_.Bytes(),
			})
		}
	}
	return arr
}

// WireToEventsByPeer converts from wire.
func WireToEventsByPeer(arr []*wire.EventDescr) EventsByPeer {
	res := EventsByPeer{}

	for _, w := range arr {
		creator := hash.BytesToPeer(w.Creator)
		h := hash.BytesToEvent(w.Hash)
		if res[creator] == nil {
			res[creator] = hash.Events{}
		}
		if !res[creator].Add(h) {
			logger.Get().Fatal("double value is detected")
		}
	}

	return res
}

/*
 * Utils:
 */

// rootZero makes roots from single event.
func rootZero(node hash.Peer) EventsByPeer {
	return EventsByPeer{
		node: hash.NewEvents(hash.ZeroEvent),
	}
}

// rootFrom makes roots from single event.
func rootFrom(e *Event) EventsByPeer {
	return EventsByPeer{
		e.Creator: hash.NewEvents(e.Hash()),
	}
}
