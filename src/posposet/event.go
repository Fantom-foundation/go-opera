package posposet

import (
	"fmt"
)

// Event is a poset event.
type Event struct {
	Creator Address
	Parents EventHashes

	hash    EventHash
	parents map[EventHash]*Event
}

// Hash calcs hash of event.
func (e *Event) Hash() EventHash {
	return EventHashOf(e)
}

// String returns string representation.
func (e *Event) String() string {
	hash := e.Hash()
	return fmt.Sprintf("Event{%s, %s}", hash.ShortString(), e.Parents.ShortString())
}
