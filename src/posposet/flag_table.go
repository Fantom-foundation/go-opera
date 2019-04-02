package posposet

import (
	"fmt"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// TODO: make FlagTable internal
// TODO: make EventsByNode internal
// TODO: cache PoS-stake at FlagTable

type (
	// EventsByNode is a event hashes groupped by creator.
	// ( creator --> event hashes )
	EventsByPeer map[hash.Peer]hash.Events

	// FlagTable stores the reachability of each event to the roots.
	// It helps to select root without using path searching algorithms.
	// Zero-hash is a self-parent root.
	// ( event hash --> root creator --> root hashes )
	FlagTable map[hash.Event]EventsByPeer
)

/*
 * FlagTable's methods:
 */

// IsRoot returns true if event is root.
func (ft FlagTable) IsRoot(event hash.Event) bool {
	if knowns := ft[event]; knowns != nil {
		for _, events := range knowns {
			if events.Contains(event) {
				return true
			}
		}
	}
	return false
}

// Roots returns all roots by node.
func (ft FlagTable) Roots() EventsByPeer {
	roots := EventsByPeer{}
	for event, events := range ft {
		for node, hashes := range events {
			if hashes.Contains(event) {
				roots.AddOne(event, node)
			}
		}
	}
	return roots
}

// EventKnows return true if e knows event of node.
func (ft FlagTable) EventKnows(e hash.Event, node hash.Peer, event hash.Event) bool {
	return ft[e] != nil && ft[e].Contains(node, event)
}

// ToWire converts to simple slice.
func (ft FlagTable) ToWire() []*wire.Flag {
	var arr []*wire.Flag
	for event, roots := range ft {
		arr = append(arr, &wire.Flag{
			Event: event.Bytes(),
			Roots: roots.ToWire(),
		})
	}
	return arr
}

// WireToFlagTable converts from wire.
func WireToFlagTable(arr []*wire.Flag) FlagTable {
	res := FlagTable{}

	for _, w := range arr {
		event := hash.BytesToEventHash(w.Event)
		res[event] = WireToEventsByPeer(w.Roots)
	}

	return res
}

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
		for hash := range hh {
			arr = append(arr, &wire.EventDescr{
				Creator: creator.Bytes(),
				Hash:    hash.Bytes(),
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
		h := hash.BytesToEventHash(w.Hash)
		if res[creator] == nil {
			res[creator] = hash.Events{}
		}
		if !res[creator].Add(h) {
			panic(fmt.Errorf("Double value is detected"))
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
