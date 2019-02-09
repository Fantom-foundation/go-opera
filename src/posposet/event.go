package posposet

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// Event is a poset event.
type Event struct {
	Creator common.Address
	Parents EventHashes

	hash    EventHash            // cache for .Hash()
	parents map[EventHash]*Event // temporary cache for internal purpose
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
	return fmt.Sprintf("Event{%s, %s}", e.Hash().ShortString(), e.Parents.ShortString())
}
