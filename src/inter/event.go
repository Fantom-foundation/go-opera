package inter

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

/*
 * Event:
 */

// Event is a poset event.
type Event struct {
	Index                uint64
	Creator              hash.Peer
	Parents              hash.Events
	LamportTime          Timestamp
	InternalTransactions []*InternalTransaction
	ExternalTransactions [][]byte
	Sign                 []byte

	hash hash.Event // cache for .Hash()
}

// Hash calcs hash of event.
func (e *Event) Hash() hash.Event {
	if e.hash.IsZero() {
		e.hash = EventHashOf(e)
	}
	return e.hash
}

// String returns string representation.
func (e *Event) String() string {
	return fmt.Sprintf("Event{%s, %s, t=%d}", e.Hash().String(), e.Parents.String(), e.LamportTime)
}

// ToWire converts to proto.Message.
func (e *Event) ToWire() *wire.Event {
	return &wire.Event{
		Index:                e.Index,
		Creator:              e.Creator.Bytes(),
		Parents:              e.Parents.ToWire(),
		LamportTime:          uint64(e.LamportTime),
		InternalTransactions: InternalTransactionsToWire(e.InternalTransactions),
		ExternalTransactions: e.ExternalTransactions,
		Sign:                 e.Sign,
	}
}

// WireToEvent converts from wire.
func WireToEvent(w *wire.Event) *Event {
	if w == nil {
		return nil
	}
	return &Event{
		Index:                w.Index,
		Creator:              hash.BytesToPeer(w.Creator),
		Parents:              hash.WireToEventHashes(w.Parents),
		LamportTime:          Timestamp(w.LamportTime),
		InternalTransactions: WireToInternalTransactions(w.InternalTransactions),
		ExternalTransactions: w.ExternalTransactions,
		Sign:                 w.Sign,
	}
}

/*
 * Utils:
 */

// EventHashOf calcs hash of event.
func EventHashOf(e *Event) hash.Event {
	w := e.ToWire()
	w.Sign = nil
	buf, err := proto.Marshal(w)
	if err != nil {
		panic(err)
	}
	return hash.Event(hash.Of(buf))
}

// FakeFuzzingEvents generates random independent events for test purpose.
func FakeFuzzingEvents() (res []*Event) {
	creators := []hash.Peer{
		hash.Peer{},
		hash.FakePeer(),
		hash.FakePeer(),
		hash.FakePeer(),
	}
	parents := []hash.Events{
		hash.FakeEvents(0),
		hash.FakeEvents(1),
		hash.FakeEvents(8),
	}
	for c := 0; c < len(creators); c++ {
		for p := 0; p < len(parents); p++ {
			e := &Event{
				Index:   uint64(c*len(parents) + p),
				Creator: creators[c],
				Parents: parents[p],
				InternalTransactions: []*InternalTransaction{
					&InternalTransaction{
						Amount:   999,
						Receiver: creators[c],
					},
				},
				ExternalTransactions: nil,
			}
			res = append(res, e)
		}
	}
	return
}
