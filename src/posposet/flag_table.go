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
	// Events its a event hashes groupped by creator.
	// ( creator --> event hashes )
	Events map[common.Address]EventHashes

	// FlagTable stores the reachability of each top event to the roots.
	// It helps to select root without using path searching algorithms.
	// Zero-hash is a self-parent root.
	// ( node --> root creator --> root hashes )
	FlagTable map[common.Address]Events

	// event is an internal struct for serialization purpose.
	event struct {
		Creator common.Address
		Hash    EventHash
	}

	// flag is an internal struct for serialization purpose.
	flag struct {
		Node  common.Address
		Roots Events
	}
)

// Add unions roots into one.
func (rr Events) Add(roots Events) (changed bool) {
	for creator, hashes := range roots {
		if rr[creator] == nil {
			rr[creator] = EventHashes{}
		}
		if rr[creator].Add(hashes.Slice()...) {
			changed = true
		}
	}
	return
}

// String returns human readable string representation.
func (rr Events) String() string {
	var ss []string
	for node, roots := range rr {
		ss = append(ss, node.String()+":"+roots.String())
	}
	return "Roots{" + strings.Join(ss, ", ") + "}"
}

// EncodeRLP is a specialized encoder to encode index into array.
func (rr Events) EncodeRLP(w io.Writer) error {
	var arr []event
	for creator, hh := range rr {
		for hash := range hh {
			arr = append(arr, event{creator, hash})
		}
	}
	return rlp.Encode(w, arr)
}

// DecodeRLP is a specialized decoder to decode index from array.
func (rr *Events) DecodeRLP(s *rlp.Stream) error {
	var arr []event
	err := s.Decode(&arr)
	if err != nil {
		return err
	}

	res := Events{}
	for _, e := range arr {
		if res[e.Creator] == nil {
			res[e.Creator] = EventHashes{}
		}
		if !res[e.Creator].Add(e.Hash) {
			return fmt.Errorf("Double value is detected")
		}
	}

	*rr = res
	return nil
}

// EncodeRLP is a specialized encoder to encode index into array.
func (ft FlagTable) EncodeRLP(w io.Writer) error {
	var arr []flag
	for node, roots := range ft {
		arr = append(arr, flag{node, roots})
	}
	return rlp.Encode(w, arr)
}

// DecodeRLP is a specialized decoder to decode index from array.
func (ft *FlagTable) DecodeRLP(s *rlp.Stream) error {
	var arr []flag
	err := s.Decode(&arr)
	if err != nil {
		return err
	}

	res := FlagTable{}
	for _, f := range arr {
		res[f.Node] = f.Roots
	}

	*ft = res
	return nil
}

/*
 * Utils:
 */

// rootZero makes roots from single event.
func rootZero(node common.Address) Events {
	return Events{
		node: newEventHashes(ZeroEventHash),
	}
}

// rootFrom makes roots from single event.
func rootFrom(e *Event) Events {
	return Events{
		e.Creator: newEventHashes(e.Hash()),
	}
}
