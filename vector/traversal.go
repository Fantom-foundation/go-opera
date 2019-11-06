package vector

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/inter"
)

type EventHeaderStack []*inter.EventHeaderData

func (s *EventHeaderStack) Pop() *inter.EventHeaderData {
	l := len(*s)
	if l == 0 {
		return nil
	}

	res := (*s)[l-1]
	*s = (*s)[:l-1]

	return res
}

func (s *EventHeaderStack) Push(v *inter.EventHeaderData) {
	*s = append(*s, v)
}

// dfsSubgraph iterates all the event which are observed by head, and accepted by a filter
func (vi *Index) dfsSubgraph(head *inter.EventHeaderData, walk func(*inter.EventHeaderData) (godeeper bool)) error {
	stack := make(EventHeaderStack, 0, vi.validators.Len())

	for next := head; next != nil; next = stack.Pop() {
		curr := *next

		// filter
		if !walk(&curr) {
			continue
		}

		// memorize parents
		for _, parent := range curr.Parents {
			p := vi.getEvent(parent)
			if p == nil {
				return errors.New("event not found " + curr.String())
			}
			stack.Push(p)
		}
	}

	return nil
}

