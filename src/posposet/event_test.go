package posposet

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

func TestEventsSort(t *testing.T) {
	assert := assert.New(t)

	expected := Events{
		&Event{consensusTime: 1, Event: inter.Event{LamportTime: 7}},
		&Event{consensusTime: 1, Event: inter.Event{LamportTime: 8}},
		&Event{consensusTime: 2, Event: inter.Event{LamportTime: 1}},
		&Event{consensusTime: 3, Event: inter.Event{LamportTime: 0}},
		&Event{consensusTime: 3, Event: inter.Event{LamportTime: 9}},
		&Event{consensusTime: 4, Event: inter.Event{LamportTime: 1}},
	}
	n := len(expected)

	for i := 0; i < 3; i++ {
		perms := rand.Perm(n)

		ordered := make(Events, n)
		for i := 0; i < n; i++ {
			ordered[i] = expected[perms[i]]
		}
		sort.Sort(ordered)

		if !assert.Equal(expected, ordered, fmt.Sprintf("perms: %v", perms)) {
			break
		}
	}
}

func TestEventsByParents(t *testing.T) {
	_, events := GenEventsByNode(5, 10, 3)
	var unordered Events
	for _, ee := range events {
		unordered = append(unordered, ee...)
	}

	ordered := unordered.ByParents()
	position := make(map[hash.Event]int)
	for i, e := range ordered {
		position[e.Hash()] = i
	}

	for i, e := range ordered {
		for p := range e.Parents {
			pos, ok := position[p]
			if !ok {
				continue
			}
			if pos > i {
				t.Fatalf("parent %s is not before %s", p.String(), e.Hash().String())
				return
			}
		}
	}
}
