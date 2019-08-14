package vector

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// dfsSubgraph returns all the event which are seen by head, and accepted by a filter
func (vi *Index) dfsSubgraph(head hash.Event, walk func(*event) (godeeper bool)) error {
	stack := make(hash.EventsStack, 0, len(vi.members))

	for next := &head; next != nil; next = stack.Pop() {
		curr := *next

		event := vi.GetEvent(curr)
		if event == nil {
			return errors.New("event wasn't found " + curr.String())
		}

		// filter
		if !walk(event) {
			continue
		}

		// memorize parents
		for _, parent := range event.Parents {
			stack.Push(parent)
		}
	}

	return nil
}
