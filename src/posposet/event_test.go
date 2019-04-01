package posposet

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

func TestEventSerialization(t *testing.T) {
	assert := assert.New(t)

	events := FakeFuzzingEvents()
	for _, e0 := range events {
		buf, err := proto.Marshal(e0.ToWire())
		assert.NoError(err)

		w := &wire.Event{}
		err = proto.Unmarshal(buf, w)
		if !assert.NoError(err) {
			break
		}
		e1 := WireToEvent(w)

		if !assert.Equal(e0, e1) {
			break
		}
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

func TestEventsByParents(t *testing.T) {
	_, events := GenEventsByNode(5, 10, 3)
	var unordered Events
	for _, ee := range events {
		unordered = append(unordered, ee...)
	}

	ordered := unordered.ByParents()
	position := make(map[hash.EventHash]int)
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

func TestEventHash(t *testing.T) {
	var (
		events = FakeFuzzingEvents()
		hashes = make([]hash.EventHash, len(events))
	)

	t.Run("Calculation", func(t *testing.T) {
		for i, e := range events {
			hashes[i] = e.Hash()
		}
	})

	t.Run("Comparison", func(t *testing.T) {
		for i, e := range events {
			h := e.Hash()
			if h != hashes[i] {
				t.Fatal("Non-deterministic event hash detected")
			}
			for _, other := range hashes[i+1:] {
				if h == other {
					t.Fatal("Event hash сollision detected")
				}
			}
		}
	})
}
