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

func Test_parents_Sum(t *testing.T) {
	type fields struct {
		cache map[hash.Event]*parent
		Mutex sync.Mutex
	}
	type args struct {
		e hash.Event
	}

	events := hash.FakeEvents(3)
	eventsArr := events.Slice()
	testEvent := hash.FakeEvent()

	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{
			name: "not found event in cache",
			args: struct{ e hash.Event }{
				e: testEvent,
			},
			fields: struct {
				cache map[hash.Event]*parent
				Mutex sync.Mutex
			}{
				cache: make(map[hash.Event]*parent),
				Mutex: sync.Mutex{}},
		},
		{
			name: "event with parents",
			args: struct{ e hash.Event }{
				e: testEvent,
			},
			fields: struct {
				cache map[hash.Event]*parent
				Mutex sync.Mutex
			}{
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
				Mutex: sync.Mutex{}},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := &parents{
				cache: tt.fields.cache,
				Mutex: tt.fields.Mutex,
			}
			if got := pp.Sum(tt.args.e); got != tt.want {
				t.Errorf("parents.Sum() = %v, want %v", got, tt.want)
			}
		})
	}
}
