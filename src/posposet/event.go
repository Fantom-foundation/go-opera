package posposet

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

/*
 * Event:
 */

// Event is a poset event.
type Event struct {
	Creator     common.Address
	Parents     EventHashes
	LamportTime Timestamp

	hash          EventHash            // cache for .Hash()
	consensusTime Timestamp            // for internal purpose
	parents       map[EventHash]*Event // TODO: move this temporary cache into Poset for root selection purpose
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

/*
 * Events:
 */

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
