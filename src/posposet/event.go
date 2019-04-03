package posposet

import (
	"bytes"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

/*
 * Event:
 */

// Event is a poset event for internal purpose.
type Event struct {
	inter.Event

	consensusTime inter.Timestamp
	parents       map[hash.Event]*Event
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
	exists := hash.Events{}
	for i, e := range ee {
		unsorted[i] = e
		exists.Add(e.Hash())
	}
	ready := hash.Events{}
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
