package posposet

import (
	"bytes"
	"sort"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

/*
 * Event:
 */

// Event is a poset event for internal purpose.
type Event struct {
	*inter.Event

	consensusTime inter.Timestamp

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

/*
 * Events:
 */

// Events is a ordered slice of events.
type Events []*Event

// String returns human readable representation.
func (ee Events) String() string {
	ss := make([]string, len(ee))
	for i := 0; i < len(ee); i++ {
		ss[i] = ee[i].String()
	}
	return strings.Join(ss, " ")
}

func (ee Events) Len() int      { return len(ee) }
func (ee Events) Swap(i, j int) { ee[i], ee[j] = ee[j], ee[i] }
func (ee Events) Less(i, j int) bool {
	a, b := ee[i], ee[j]
	return (a.consensusTime < b.consensusTime) ||
		(a.consensusTime == b.consensusTime && (a.LamportTime < b.LamportTime ||
			a.LamportTime == b.LamportTime && bytes.Compare(a.Hash().Bytes(), b.Hash().Bytes()) < 0))
}
