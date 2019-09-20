package poset

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type eventFilterFn func(event *inter.EventHeaderData) bool

// dfsSubgraph returns all the event which are caused by head, and accepted by a filter.
func (p *Poset) dfsSubgraph(head hash.Event, filter eventFilterFn) error {
	stack := make(hash.EventsStack, 0, len(p.Members) * 10)

	for pwalk := &head; pwalk != nil; pwalk = stack.Pop() {
		walk := *pwalk

		event := p.input.GetEventHeader(p.EpochN, walk)
		if event == nil {
			return errors.New("event wasn't found " + walk.String())
		}

		// filter
		if !filter(event) {
			continue
		}

		// memorize parents
		for _, parent := range event.Parents {
			stack.Push(parent)
		}
	}

	return nil
}
