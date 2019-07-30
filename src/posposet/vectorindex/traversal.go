package vectorindex

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/utils"
)

// return false to prevent walking to parents
type eventCallbackFn func(event *Event) bool

// dfsSubgraph returns all the event which are seen by head, and accepted by a filter
func (vi *Vindex) dfsSubgraph(head hash.Event, callback eventCallbackFn) error {
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
		for _, parent := range event.Parents {
			stack.Push(parent)
		}
	}

	return nil
}
