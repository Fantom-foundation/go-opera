package posposet

import (
	"errors"
	"github.com/Fantom-foundation/go-lachesis/src/utils"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type eventFilterFn func(event *inter.Event) bool

// dfsSubgraph returns all the event which are seen by head, and accepted by a filter
func (p *Poset) dfsSubgraph(head hash.Event, filter eventFilterFn) (res inter.Events, err error) {
	res = make(inter.Events, 0, 1024)

	visited := make(map[hash.Event]bool)
	stack := make(utils.EventHashesStack, 0, len(p.members))

	for pwalk := &head; pwalk != nil; pwalk = stack.Pop() {
		// ensure visited once
		walk := *pwalk
		if visited[walk] {
			continue
		}
		visited[walk] = true

		event := p.input.GetEvent(walk)
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
