package posposet

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

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
