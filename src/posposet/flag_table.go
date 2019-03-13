package posposet

import (
	"fmt"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// TODO: make FlagTable internal
// TODO: cache PoS-stake at FlagTable

type (
	// eventsByNode is a event hashes groupped by creator.
	// ( creator --> event hashes )
	eventsByNode map[common.Address]EventHashes

	// FlagTable stores the reachability of each event to the roots.
	// It helps to select root without using path searching algorithms.
	// Zero-hash is a self-parent root.
	// ( event hash --> root creator --> root hashes )
	FlagTable map[EventHash]eventsByNode
)

/*
 * FlagTable's methods:
 */

// IsRoot returns true if event is root.
func (ft FlagTable) IsRoot(event EventHash) bool {
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
func (ft FlagTable) Roots() eventsByNode {
	roots := eventsByNode{}
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
func (ft FlagTable) EventKnows(e EventHash, node common.Address, event EventHash) bool {
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
		event := BytesToEventHash(w.Event)
		res[event] = WireToEventsByNode(w.Roots)
	}

	return res
}

/*
 * eventsByNode's methods:
 */

// Add unions events into one.
func (ee eventsByNode) Add(events eventsByNode) (changed bool) {
	for creator, hashes := range events {
		if ee[creator] == nil {
			ee[creator] = EventHashes{}
		}
		if ee[creator].Add(hashes.Slice()...) {
			changed = true
		}
	}
	return
}

// AddOne appends one event.
func (ee eventsByNode) AddOne(event EventHash, creator common.Address) (changed bool) {
	if ee[creator] == nil {
		ee[creator] = EventHashes{}
	}
	if ee[creator].Add(event) {
		changed = true
	}
	return
}

// Contains returns true if event of node is in.
func (ee eventsByNode) Contains(node common.Address, event EventHash) bool {
	return ee[node] != nil && ee[node].Contains(event)
}

// Add unions roots into one.
func (ee eventsByNode) Each() map[EventHash]common.Address {
	res := make(map[EventHash]common.Address)
	for creator, events := range ee {
		for e := range events {
			res[e] = creator
		}
	}
	return res
}

// String returns human readable string representation.
func (ee eventsByNode) String() string {
	var ss []string
	for node, roots := range ee {
		ss = append(ss, node.String()+":"+roots.String())
	}
	return "byNode{" + strings.Join(ss, ", ") + "}"
}

// ToWire converts to simple slice.
func (ee eventsByNode) ToWire() []*wire.EventDescr {
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

// WireToEventsByNode converts from wire.
func WireToEventsByNode(arr []*wire.EventDescr) eventsByNode {
	res := eventsByNode{}

	for _, w := range arr {
		creator := common.BytesToAddress(w.Creator)
		hash := BytesToEventHash(w.Hash)
		if res[creator] == nil {
			res[creator] = EventHashes{}
		}
		if !res[creator].Add(hash) {
			panic(fmt.Errorf("Double value is detected"))
		}
	}

	return res
}

/*
 * Utils:
 */

// rootZero makes roots from single event.
func rootZero(node common.Address) eventsByNode {
	return eventsByNode{
		node: newEventHashes(ZeroEventHash),
	}
}

// rootFrom makes roots from single event.
func rootFrom(e *Event) eventsByNode {
	return eventsByNode{
		e.Creator: newEventHashes(e.Hash()),
	}
}
