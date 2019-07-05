package posposet

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type eventsStack []hash.Event

func (s eventsStack) Push(v hash.Event) eventsStack {
	return append(s, v)
}

func (s eventsStack) Pop() (eventsStack, *hash.Event) {
	l := len(s)
	if l == 0 {
		return s, nil
	}
	return  s[:l-1], &s[l-1]
}

type eventFilterFn func(event *inter.Event) bool

// Depth First Search
// @return all the event which are seen by head, and accepted by a filter
func (p *Poset) dfsSubgraph(head hash.Event, filter eventFilterFn) (res inter.Events, err error) {
	res = make(inter.Events, 0, 1024)

	visited := make(map[hash.Event]bool)
	stack := make(eventsStack, 0, len(p.members))

	for pwalk := &head; pwalk != nil; stack, pwalk = stack.Pop() {
		// ensure visited once
		walk := *pwalk
		if visited[walk] {
			continue
		}
		visited[walk] = true

		// get event
		if !p.input.HasEvent(walk) {
			return nil, errors.New("event wasn't found " + walk.String())
		}
		event := p.input.GetEvent(walk)

		// filter
		if !filter(event) {
			continue
		}

		// collect event
		res = append(res, event)

		// memorize parents
		for parent := range event.Parents {
			stack = stack.Push(parent)
		}
	}

	return res, nil
}
