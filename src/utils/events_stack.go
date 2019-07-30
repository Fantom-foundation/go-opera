package utils

import "github.com/Fantom-foundation/go-lachesis/src/hash"

type EventHashesStack []hash.Event

func (s *EventHashesStack) Push(v hash.Event) {
	*s = append(*s, v)
}

func (s *EventHashesStack) Pop() *hash.Event {
	l := len(*s)
	if l == 0 {
		return nil
	}

	res := &(*s)[l-1]
	*s = (*s)[:l-1]

	return res
}
