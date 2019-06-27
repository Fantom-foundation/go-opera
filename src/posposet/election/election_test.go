package election

import (
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/ordering"
)

type fakeEdge struct {
	from hash.Event
	to   RootSlot
}

type (
	stakes map[string]*big.Int
)

type testExpected struct {
	DecidedFrame     FrameHeight
	DecidedSfWitness string
	DecisiveRoots    map[string]bool
}

func TestProcessRoot(t *testing.T) {

	t.Run("4 uni notDecided", func(t *testing.T) {
		testProcessRoot(t,
			nil,
			stakes{
				"nodeA": big.NewInt(1),
				"nodeB": big.NewInt(1),
				"nodeC": big.NewInt(1),
				"nodeD": big.NewInt(1),
			}, `
a0_0  b0_0  c0_0  d0_0
║     ║     ║     ║
a1_1══╬═════╣     ║
║     ║     ║     ║
║╚════b1_1══╣     ║
║     ║     ║     ║
║     ║╚════c1_1══╣
║     ║     ║     ║
║     ║╚═══─╫╩════d1_1
║     ║     ║     ║
a2_2══╬═════╬═════╣
║     ║     ║     ║
`)
	})

	t.Run("4 equalStakes", func(t *testing.T) {
		testProcessRoot(t,
			&testExpected{
				DecidedFrame:     0,
				DecidedSfWitness: "c0_0",
				DecisiveRoots:    map[string]bool{"a2_2": true},
			},
			stakes{
				"nodeA": big.NewInt(1),
				"nodeB": big.NewInt(1),
				"nodeC": big.NewInt(1),
				"nodeD": big.NewInt(1),
			}, `
a0_0  b0_0  c0_0  d0_0
║     ║     ║     ║
a1_1══╬═════╣     ║
║     ║     ║     ║
║     b1_1══╬═════╣
║     ║     ║     ║
║     ║╚════c1_1══╣
║     ║     ║     ║
║     ║╚═══─╫╩════d1_1
║     ║     ║     ║
a2_2══╬═════╬═════╣
║     ║     ║     ║
`)
	})

	t.Run("4 equalStakes missingRoot", func(t *testing.T) {
		testProcessRoot(t,
			&testExpected{
				DecidedFrame:     0,
				DecidedSfWitness: "c0_0",
				DecisiveRoots:    map[string]bool{"a2_2": true},
			},
			stakes{
				"nodeA": big.NewInt(1),
				"nodeB": big.NewInt(1),
				"nodeC": big.NewInt(1),
				"nodeD": big.NewInt(1),
			}, `
a0_0  b0_0  c0_0  d0_0
║     ║     ║     ║
a1_1══╬═════╣     ║
║     ║     ║     ║
║╚════b1_1══╣     ║
║     ║     ║     ║
║╚═══─╫╩════c1_1  ║
║     ║     ║     ║
a2_2══╬═════╣     ║
║     ║     ║     ║
`)
	})

	t.Run("4 differentStakes", func(t *testing.T) {
		testProcessRoot(t,
			&testExpected{
				DecidedFrame:     0,
				DecidedSfWitness: "c0_0",
				DecisiveRoots:    map[string]bool{"b2_2": true},
			},
			stakes{
				"nodeA": big.NewInt(1000000000000000000),
				"nodeB": big.NewInt(1),
				"nodeC": big.NewInt(1),
				"nodeD": big.NewInt(1),
			}, `
a0_0  b0_0  c0_0  d0_0
║     ║     ║     ║
a1_1══╬═════╣     ║
║     ║     ║     ║
║╚════+b1_1 ║     ║
║     ║     ║     ║
║╚═══─╫─════+c1_1 ║
║     ║     ║     ║
║╚═══─╫╩═══─╫╩════d1_1
║     ║     ║     ║
╠═════b2_2══╬═════╣
║     ║     ║     ║
`)
	})

	t.Run("4 differentStakes 4rounds", func(t *testing.T) {
		testProcessRoot(t,
			&testExpected{
				DecidedFrame:     0,
				DecidedSfWitness: "b0_0",
				DecisiveRoots:    map[string]bool{"a4_4": true},
			},
			stakes{
				"nodeA": big.NewInt(4),
				"nodeB": big.NewInt(2),
				"nodeC": big.NewInt(1),
				"nodeD": big.NewInt(1),
			}, `
a0_0  b0_0  c0_0  d0_0
║     ║     ║     ║
a1_1══╣     ║     ║
║     ║     ║     ║
║     +b1_1═╬═════╣
║     ║     ║     ║
║╚═══─╫─════c1_1══╣
║     ║     ║     ║
║╚═══─╫─═══─╫╩════d1_1
║     ║     ║     ║
a2_2  ╣     ║     ║
║     ║     ║     ║
║╚════b2_2══╬═════╣
║     ║     ║     ║
║╚═══─╫╩════c2_2══╣
║     ║     ║     ║
║╚═══─╫╩═══─╫─════+d2_2
║     ║     ║     ║
a3_3══╬═════╬═════╣
║     ║     ║     ║
║╚════b3_3══╬═════╣
║     ║     ║     ║
║╚═══─╫╩════c3_3══╣
║     ║     ║     ║
║╚═══─╫╩═══─╫╩════d3_3
║     ║     ║     ║
a4_4══╣     ║     ║
║     ║     ║     ║
`)
	})

}

func testProcessRoot(
	t *testing.T,
	expected *testExpected,
	stakes stakes,
	dag string,
) {
	assertar := assert.New(t)

	peers, _, named := inter.ASCIIschemeToDAG(dag)

	// nodes:
	totalStake := new(big.Int)
	nodes := make([]ElectionNode, 0, len(peers))
	for _, peer := range peers {
		n := ElectionNode{
			Nodeid:      peer,
			StakeAmount: stakes[peer.String()],
		}
		nodes = append(nodes, n)
		totalStake.Add(totalStake, n.StakeAmount)
	}

	superMajority := get2of3(totalStake)
	//t.Logf("superMajority = %s", superMajority.String())

	// events:
	events := make(map[hash.Event]*inter.Event)
	vertices := make(map[hash.Event]RootSlot)
	edges := make(map[fakeEdge]hash.Event)

	for dsc, root := range named {
		events[root.Hash()] = root
		h := root.Hash()

		vertices[h] = RootSlot{
			Frame:  frameOf(dsc),
			Nodeid: root.Creator,
		}
	}

	for dsc, root := range named {
		noPrev := false
		if strings.HasPrefix(dsc, "+") {
			noPrev = true
			dsc = strings.TrimPrefix(dsc, "+")
		}
		from := root.Hash()
		for sSeen := range root.Parents {
			if sSeen.IsZero() {
				continue
			}
			if p := events[sSeen]; p.Creator == root.Creator && noPrev {
				continue
			}
			to := sSeen
			edge := fakeEdge{
				from: from,
				to:   vertices[to],
			}
			edges[edge] = to
		}
	}

	// strongly see fn:
	stronglySeeFn := func(a hash.Event, b RootSlot) *hash.Event {
		edge := fakeEdge{
			from: a,
			to:   b,
		}
		hashB, ok := edges[edge]
		if ok {
			return &hashB
		} else {
			return nil
		}
	}

	election := NewElection(nodes, totalStake, superMajority, 0, stronglySeeFn)

	// ordering:
	var (
		processed = make(map[hash.Event]*inter.Event)
		alreadyDecided = false
	)
	orderThenProcess := ordering.EventBuffer(ordering.Callback{

		Process: func(root *inter.Event) {
			rootHash := root.Hash()
			rootSlot, ok := vertices[rootHash]
			if !ok {
				t.Fatal("inconsistent vertices")
			}
			got, err := election.ProcessRoot(rootHash, rootSlot)
			if err != nil {
				t.Fatal(err)
			}
			processed[root.Hash()] = root

			// checking:
			decisive := expected != nil && expected.DecisiveRoots[root.Hash().String()]
			if decisive || alreadyDecided {
				assertar.NotNil(got)
				assertar.Equal(expected.DecidedFrame, got.DecidedFrame)
				assertar.Equal(expected.DecidedSfWitness, got.DecidedSfWitness.String())
				alreadyDecided = true
			} else {
				assertar.Nil(got)
			}
		},

		Drop: func(e *inter.Event, err error) {
			t.Fatal(e, err)
		},

		Exists: func(h hash.Event) *inter.Event {
			return processed[h]
		},
	})

	// processing:
	for _, root := range named {
		orderThenProcess(root)
	}
}

func get2of3(x *big.Int) *big.Int {
	res := new(big.Int)
	res.
		Mul(x, big.NewInt(2)).
		Div(res, big.NewInt(3)).
		Add(res, big.NewInt(1))

	return res
}

func frameOf(dsc string) FrameHeight {
	s := strings.Split(dsc, "_")[1]
	h, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return FrameHeight(h)
}
