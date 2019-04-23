package posnode

import (
	"fmt"
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

	for h, e := range node.parents.cache {
		fmt.Printf("====> %s: %+v (%f)\n", h, e, node.parents.Sum(h))
	}

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
