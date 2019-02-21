package posposet

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// Event is a poset event.
type Event struct {
	Creator     common.Address
	Parents     EventHashes
	LamportTime uint64

	hash    EventHash            // cache for .Hash()
	parents map[EventHash]*Event // TODO: move this temporary cache into Poset for root selection purpose
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
	return fmt.Sprintf("Event{%s, %s, %d}", e.Hash().String(), e.Parents.String(), e.LamportTime)
}
