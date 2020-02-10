package poset

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
)

type eventFilterFn func(event *inter.EventHeaderData) bool

// dfsSubgraph iterates all the events which are observed by head, and accepted by a filter.
// filter MAY BE called twice for the same event.
func (p *Poset) dfsSubgraph(head hash.Event, filter eventFilterFn) error {
	stack := make(hash.EventsStack, 0, p.Validators.Len()*10)

	for pwalk := &head; pwalk != nil; pwalk = stack.Pop() {
		walk := *pwalk

		event := p.input.GetEventHeader(p.EpochN, walk)
		if event == nil {
			return errors.New("event not found " + walk.String())
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
