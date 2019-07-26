package posposet

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type eventsStack []hash.Event

func (s *eventsStack) Push(v hash.Event) {
	*s = append(*s, v)
}

func (s *eventsStack) Pop() *hash.Event {
	l := len(*s)
	if l == 0 {
		return nil
	}

	res := &(*s)[l-1]
	*s = (*s)[:l-1]

	return res
}

type eventFilterFn func(event *inter.Event) bool

// dfsSubgraph returns all the event which are seen by head, and accepted by a filter
func (p *Poset) dfsSubgraph(head hash.Event, filter eventFilterFn) (res inter.Events, err error) {
	res = make(inter.Events, 0, 1024)

	visited := make(map[hash.Event]bool)
	stack := make(eventsStack, 0, len(p.members))

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
		for parent := range event.Parents {
			if !parent.IsZero() {
				stack.Push(parent)
			}
		}
	}

	return res, nil
}
