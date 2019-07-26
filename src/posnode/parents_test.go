package posnode

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

func TestParentSelection(t *testing.T) {
	logger.SetTestMode(t)

	testParentSelection(t, "empty", `
`)

	testParentSelection(t, "single", `
1:A1
`)

	testParentSelection(t, "simple", `
1:a1   1:b1   1:c1
║      ║      ║
1:a2 ─ ╬ ─ ─  ╣      1:d1
║      ║      ║      ║
║      ╠ ─ ─  1:c2 ─ ╣
║      ║      ║      ║
╠ ─ ─  1:b2 ─ ╣      ║
║      ║      ║      ║
╠ ─ ─  ╬ ─ ─  1:c3   ║
║      ║      ║      ║
║      ╠ ─ ─  ╫ ─ ─  1:D2
║      ║      ║      ║
`)
}

// testParentSelection uses special event name format: <creator weight>:<uppercase if expected|lowercase><expected ordering num>
func testParentSelection(t *testing.T, dsc, schema string) {
	t.Run(dsc, func(t *testing.T) {
		assertar := assert.New(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		consensus := NewMockConsensus(ctrl)
		consensus.EXPECT().
			CurrentSuperFrameN().
			Return(idx.SuperFrame(1)).
			AnyTimes()
		consensus.EXPECT().
			PushEvent(gomock.Any()).
			AnyTimes()

		store := NewMemStore()
		node := NewForTests(dsc, store, consensus)
		node.initParents()
		defer node.Stop()

		expected := ASCIIschemeToDAG(node, consensus, schema)

		for n, expect := range expected {
			parent := node.parents.PopBest()
			if !assertar.NotNil(parent, "step %d", n) {
				break
			}
			if !assertar.Equal(expect, parent.String(), "step %d of %v", n, expected) {
				break
			}
		}

		assertar.Nil(node.parents.PopBest(), "last step")

	})
}

func ASCIIschemeToDAG(n *Node, c *MockConsensus, schema string) (expected []string) {
	_, _, events := inter.ASCIIschemeToDAG(schema, func(e *inter.Event, nn []hash.Peer) {
		e.SfNum = c.CurrentSuperFrameN()
	})

	weights := make(map[hash.Peer]inter.Stake)
	for name, e := range events {
		w, o := parseSpecName(name)
		weights[e.Creator] = w
		if o > 0 {
			expected = append(expected, name)
		}
	}

	c.EXPECT().
		StakeOf(gomock.Any()).
		DoAndReturn(func(p hash.Peer) inter.Stake {
			return weights[p]
		}).
		AnyTimes()

	for _, e := range events {
		n.onNewEvent(e)
	}

	sort.Sort(byOrderNum(expected))
	return
}

func parseSpecName(name string) (weight inter.Stake, orderNum int64) {
	ss := strings.Split(name, ":")
	if len(ss) != 2 {
		panic("invalid event name format")
	}

	w, err := strconv.ParseUint(ss[0], 10, 64)
	if err != nil {
		panic("invalid event name format (weight): " + err.Error())
	}
	weight = inter.Stake(w)

	if ss[1][0] == strings.ToUpper(ss[1])[0] {
		orderNum, err = strconv.ParseInt(ss[1][1:], 10, 64)
		if err != nil {
			panic("invalid event name format (order num): " + err.Error())
		}
	} else {
		orderNum = 0
	}

	return
}

type byOrderNum []string

func (ss byOrderNum) Len() int { return len(ss) }

func (ss byOrderNum) Swap(i, j int) { ss[i], ss[j] = ss[j], ss[i] }

func (ss byOrderNum) Less(i, j int) bool {
	_, a := parseSpecName(ss[i])
	_, b := parseSpecName(ss[j])

	return a < b
}
