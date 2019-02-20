package posposet

import (
	"fmt"
	"io"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/rlp"
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

	// storedEvent is an internal struct for serialization purpose.
	storedEvent struct {
		Creator common.Address
		Hash    EventHash
	}

	// storedFlag is an internal struct for serialization purpose.
	storedFlag struct {
		Event EventHash
		Roots eventsByNode
	}
)

/*
 * Events's methods:
 */

// Add unions roots into one.
func (ee eventsByNode) Add(roots eventsByNode) (changed bool) {
	for creator, hashes := range roots {
		if ee[creator] == nil {
			ee[creator] = EventHashes{}
		}
		if ee[creator].Add(hashes.Slice()...) {
			changed = true
		}
	}
	return
}

// String returns human readable string representation.
func (ee eventsByNode) String() string {
	var ss []string
	for node, roots := range ee {
		ss = append(ss, node.String()+":"+roots.String())
	}
	return "byNode{" + strings.Join(ss, ", ") + "}"
}

// EncodeRLP is a specialized encoder to encode index into array.
func (ee eventsByNode) EncodeRLP(w io.Writer) error {
	var arr []storedEvent
	for creator, hh := range ee {
		for hash := range hh {
			arr = append(arr, storedEvent{creator, hash})
		}
	}
	return rlp.Encode(w, arr)
}

// DecodeRLP is a specialized decoder to decode index from array.
func (ee *eventsByNode) DecodeRLP(s *rlp.Stream) error {
	var arr []storedEvent
	err := s.Decode(&arr)
	if err != nil {
		return err
	}

	res := eventsByNode{}
	for _, e := range arr {
		if res[e.Creator] == nil {
			res[e.Creator] = EventHashes{}
		}
		if !res[e.Creator].Add(e.Hash) {
			return fmt.Errorf("Double value is detected")
		}
	}

	*ee = res
	return nil
}

/*
 * FlagTable's methods:
 */

// EncodeRLP is a specialized encoder to encode index into array.
func (ft FlagTable) EncodeRLP(w io.Writer) error {
	var arr []storedFlag
	for event, roots := range ft {
		arr = append(arr, storedFlag{event, roots})
	}
	return rlp.Encode(w, arr)
}

// DecodeRLP is a specialized decoder to decode index from array.
func (ft *FlagTable) DecodeRLP(s *rlp.Stream) error {
	var arr []storedFlag
	err := s.Decode(&arr)
	if err != nil {
		return err
	}

	res := FlagTable{}
	for _, f := range arr {
		res[f.Event] = f.Roots
	}

	*ft = res
	return nil
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
