package inter

import (
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

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
