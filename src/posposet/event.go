package posposet

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

/*
 * Event:
 */

// Event is a poset event.
type Event struct {
	Index                uint64
	Creator              common.Address
	Parents              EventHashes
	LamportTime          Timestamp
	InternalTransactions []*InternalTransaction
	ExternalTransactions [][]byte

	hash          EventHash            // cache for .Hash()
	consensusTime Timestamp            // for internal purpose
	parents       map[EventHash]*Event // for internal purpose
}

// Hash calcs hash of event.
func (e *Event) Hash() EventHash {
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
	}
}

// WireToEvent converts from wire.
func WireToEvent(w *wire.Event) *Event {
	if w == nil {
		return nil
	}
	return &Event{
		Index:                w.Index,
		Creator:              common.BytesToAddress(w.Creator),
		Parents:              WireToEventHashes(w.Parents),
		LamportTime:          Timestamp(w.LamportTime),
		InternalTransactions: WireToInternalTransactions(w.InternalTransactions),
		ExternalTransactions: w.ExternalTransactions,
	}
}

/*
 * Events:
 */

// Events is a ordered slice of events.
type Events []*Event

// String returns string representation.
func (ee Events) String() string {
	ss := make([]string, len(ee))
	for i := 0; i < len(ee); i++ {
		ss[i] = ee[i].String()
	}
	return strings.Join(ss, "  ")
}

func (ee Events) Len() int      { return len(ee) }
func (ee Events) Swap(i, j int) { ee[i], ee[j] = ee[j], ee[i] }
func (ee Events) Less(i, j int) bool {
	a, b := ee[i], ee[j]
	return false ||
		(a.consensusTime < b.consensusTime) ||
		(a.consensusTime == b.consensusTime && a.LamportTime < b.LamportTime) ||
		(a.consensusTime == b.consensusTime && a.LamportTime == b.LamportTime && bytes.Compare(a.Hash().Bytes(), b.Hash().Bytes()) < 0)
}

// ByParents returns events topologically ordered by parent dependency.
// TODO: use Topological sort algorithm
func (ee Events) ByParents() (res Events) {
	unsorted := make(Events, len(ee))
	exists := EventHashes{}
	for i, e := range ee {
		unsorted[i] = e
		exists.Add(e.Hash())
	}
	ready := EventHashes{}
	for len(unsorted) > 0 {
	EVENTS:
		for i, e := range unsorted {

			for p := range e.Parents {
				if exists.Contains(p) && !ready.Contains(p) {
					continue EVENTS
				}
			}

			res = append(res, e)
			unsorted = append(unsorted[0:i], unsorted[i+1:]...)
			ready.Add(e.Hash())
			break
		}
	}

	return
}
