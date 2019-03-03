package posposet

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/rlp"
)

func TestEventSerialization(t *testing.T) {
	assert := assert.New(t)

	events := FakeFuzzingEvents()
	for _, e0 := range events {
		buf, err := rlp.EncodeToBytes(e0)
		assert.NoError(err)

		e1 := &Event{}
		err = rlp.DecodeBytes(buf, e1)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(e0, e1)
	}
}

func TestEventsSort(t *testing.T) {
	assert := assert.New(t)

	expected := Events{
		&Event{consensusTime: 1, LamportTime: 7},
		&Event{consensusTime: 1, LamportTime: 8},
		&Event{consensusTime: 2, LamportTime: 1},
		&Event{consensusTime: 3, LamportTime: 0},
		&Event{consensusTime: 3, LamportTime: 9},
		&Event{consensusTime: 4, LamportTime: 1},
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
