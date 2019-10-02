package vector

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// dfsSubgraph returns all the event which are observed by head, and accepted by a filter
func (vi *Index) dfsSubgraph(head hash.Event, walk func(*inter.EventHeaderData) (godeeper bool)) error {
	stack := make(hash.EventsStack, 0, len(vi.validators))

	for next := &head; next != nil; next = stack.Pop() {
		curr := *next

		event := vi.getEvent(curr)
		if event == nil {
			return errors.New("event not found " + curr.String())
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
