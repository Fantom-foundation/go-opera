package vector

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/utils"
)

// eventCallbackFn returns false to prevent walking to parents.
type eventCallbackFn func(e *event) bool

// dfsSubgraph returns all the event which are seen by head, and accepted by a filter
func (vi *Index) dfsSubgraph(head hash.Event, callback eventCallbackFn) error {
	stack := make(utils.EventHashesStack, 0, len(vi.members))

	for pwalk := &head; pwalk != nil; pwalk = stack.Pop() {
		walk := *pwalk

		event, ok := vi.events[walk]
		if !ok {
			return errors.New("event wasn't found " + walk.String())
		}

		// filter
		if !callback(event) {
			continue
		}

		// memorize parents
		for parent := range event.Parents {
			if !parent.IsZero() {
				stack.Push(parent)
			}
		}
	}

	return nil
}
