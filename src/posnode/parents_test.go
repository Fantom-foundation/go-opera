package posnode

import (
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// TODO: it is a stub. Test the all situations.
func TestParents(t *testing.T) {
	assert := assert.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	consensus := NewMockConsensus(ctrl)

	store := NewMemStore()
	node := NewForTests("n0", store, consensus)
	node.initParents()

	prepareParents(node, consensus, `
a01   b01   c01
║     ║     ║
a02 ─ ╬ ─ ─ ╣     d01
║     ║     ║     ║
║     ╠ ─ ─ c02 ─ ╣
║     ║     ║     ║
╠ ─ ─ b02 ─ ╣     ║
║     ║     ║     ║
╠ ─ ─ ╬ ─ ─ c03   ║
║     ║     ║     ║
║     ╠ ─ ─ ╫ ─ ─ d02
║     ║     ║     ║
`, map[string]float64{"*": 1})

	for n, expect := range []string{"c03", "d02", ""} {
		parent := node.popBestParent()
		if parent == nil {
			assert.Equal("", expect, "step %d", n)
			break
		}
		if !assert.Equal(expect, parent.String(), "step %d", n) {
			break
		}
	}
}

func prepareParents(n *Node, c *MockConsensus, schema string, stakes map[string]float64) {
	c.EXPECT().
		PushEvent(gomock.Any()).
		AnyTimes()

	c.EXPECT().
		GetStakeOf(gomock.Any()).
		DoAndReturn(func(p hash.Peer) float64 {

			if stake, ok := stakes[p.String()]; ok {
				return stake
			}
			if stake, ok := stakes["*"]; ok {
				return stake
			}
			return float64(0)

		}).
		AnyTimes()

	_, _, events := inter.ParseEvents(schema)
	unordered := make(inter.Events, 0, len(events))
	for _, e := range events {
		unordered = append(unordered, e)
	}
	for _, e := range unordered.ByParents() {
		n.saveNewEvent(e)
	}
}

func TestParentsSum(t *testing.T) {
	testEvent := hash.FakeEvent()

	t.Run("not found event in cache", func(t *testing.T) {
		pp := &parents{
			cache: make(map[hash.Event]*parent),
			Mutex: sync.Mutex{},
		}

		assert.Equal(t, pp.Sum(testEvent), float64(0))
	})

	t.Run("event with parents", func(t *testing.T) {
		events := hash.FakeEvents(3)
		eventsArr := events.Slice()
		pp := &parents{
			cache: map[hash.Event]*parent{
				testEvent: {
					Value:   1,
					Parents: events,
				},
				eventsArr[0]: {
					Value: 2,
				},
				eventsArr[1]: {
					Value: 3,
				},
				eventsArr[2]: {
					Value: 4,
				},
			},
			Mutex: sync.Mutex{},
		}

		assert.Equal(t, pp.Sum(testEvent), float64(10))
	})
}

func TestParentsDel(t *testing.T) {
	deletedEvent := hash.FakeEvent()

	t.Run("not found event", func(t *testing.T) {
		pp := &parents{
			cache: make(map[hash.Event]*parent),
			Mutex: sync.Mutex{},
		}

		assert.NotPanics(t, func() {
			pp.Del(deletedEvent)
		})
	})

	t.Run("event with parents", func(t *testing.T) {
		events := hash.FakeEvents(3)
		eventsArr := events.Slice()
		pp := &parents{
			cache: map[hash.Event]*parent{
				deletedEvent: {
					Parents: events,
				},
				eventsArr[0]: new(parent),
				eventsArr[1]: new(parent),
				eventsArr[2]: new(parent),
			},
			Mutex: sync.Mutex{},
		}

		assert.NotPanics(t, func() {
			pp.Del(deletedEvent)
		})
		assert.NotContains(t, pp.cache, deletedEvent)
	})

	t.Run("not delete other events", func(t *testing.T) {
		protectedEvent := hash.FakeEvent()
		pp := &parents{
			cache: map[hash.Event]*parent{
				deletedEvent:   new(parent),
				protectedEvent: new(parent),
			},
		}

		assert.NotPanics(t, func() {
			pp.Del(deletedEvent)
		})
		assert.NotContains(t, pp.cache, deletedEvent)
		assert.Contains(t, pp.cache, protectedEvent)
	})
}

func TestNodePopBestParent(t *testing.T) {
	t.Run("not found event", func(t *testing.T) {
		node := &Node{
			parents: parents{
				cache: make(map[hash.Event]*parent),
				Mutex: sync.Mutex{},
			},
		}

		result := node.popBestParent()
		assert.Nil(t, result)
	})

	t.Run("not the last parent with other parents", func(t *testing.T) {
		event := hash.FakeEvent()
		node := &Node{
			parents: parents{
				cache: map[hash.Event]*parent{
					hash.FakeEvent(): {
						Last: false,
					},
					event: {
						Last:  true,
						Value: float64(123),
					},
					hash.FakeEvent(): {
						Last:  true,
						Value: float64(1),
					},
				},
				Mutex: sync.Mutex{},
			},
		}

		result := node.popBestParent()
		assert.EqualValues(t, result, &event)
		assert.NotContains(t, node.parents.cache, event)
	})

	t.Run("first maximum return", func(t *testing.T) {
		event := hash.FakeEvent()
		node := &Node{
			parents: parents{
				cache: map[hash.Event]*parent{
					hash.FakeEvent(): {
						Last:  true,
						Value: float64(1),
					},
					event: {
						Last:  true,
						Value: float64(123),
					},
					hash.FakeEvent(): {
						Last:  true,
						Value: float64(123),
					},
					hash.FakeEvent(): {
						Last:  true,
						Value: float64(123),
					},
				},
				Mutex: sync.Mutex{},
			},
		}

		result := node.popBestParent()
		assert.EqualValues(t, result, &event)
		assert.NotContains(t, node.parents.cache, event)
	})
}
