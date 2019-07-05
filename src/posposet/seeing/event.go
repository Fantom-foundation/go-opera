package seeing

import (
	"bytes"
	"sort"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// Event is a poset event for internal purpose.
type Event struct {
	*inter.Event

	CreatorSeq          int64
	FirstDescendantsSeq []int64 // 0 <= FirstDescendantsSeq[i] <= 9223372036854775807
	LastAncestorsSeq    []int64 // -1 <= LastAncestorsSeq[i] <= 9223372036854775807
	FirstDescendants    []hash.Event
	LastAncestors       []hash.Event
}

// SelfParent returns self parent from event. If it returns "false" then a self
// parent is missing.
func (e *Event) SelfParent() (hash.Event, bool) {
	for parent := range e.Parents {
		if bytes.Equal(e.Creator.Bytes(), parent.Bytes()) {
			return parent, true
		}
	}

	return hash.Event{}, false
}

// OtherParents returns "other parents" sorted slice.
func (e *Event) OtherParents() hash.EventsSlice {
	parents := e.Parents.Copy()

	sp, ok := e.SelfParent()
	if ok {
		delete(parents, sp)
	}

	events := parents.Slice()
	sort.Sort(events)

	return events
}
