package poset

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type eventFilterFn func(event *inter.EventHeaderData) bool

// dfsSubgraph returns all the event which are seen by head, and accepted by a filter
func (p *Poset) dfsSubgraph(head hash.Event, filter eventFilterFn) (res []*inter.EventHeaderData, err error) {
	res = make([]*inter.EventHeaderData, 0, 1024)

	stack := make(hash.EventsStack, 0, len(p.Members))

	for pwalk := &head; pwalk != nil; pwalk = stack.Pop() {
		walk := *pwalk

		event := p.input.GetEventHeader(p.SuperFrameN, walk)
		if event == nil {
			return nil, errors.New("event wasn't found " + walk.String())
		}

		// filter
		if !filter(event) {
			continue
		}

		// collect event
		res = append(res, event)

		// memorize parents
		for _, parent := range event.Parents {
			stack.Push(parent)
		}
	}

	return res, nil
}
