package election

import (
	"github.com/Fantom-foundation/go-lachesis/src/posposet/members"
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
	stakes map[string]Amount
)

type testExpected struct {
	DecidedFrame     IdxFrame
	DecidedSfWitness string
	DecisiveRoots    map[string]bool
}

func TestProcessRoot(t *testing.T) {

	t.Run("4 equalStakes notDecided", func(t *testing.T) {
		testProcessRoot(t,
			nil,
			stakes{
				"nodeA": 1,
				"nodeB": 1,
				"nodeC": 1,
				"nodeD": 1,
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
				"nodeA": 1,
				"nodeB": 1,
				"nodeC": 1,
				"nodeD": 1,
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
				"nodeA": 1,
				"nodeB": 1,
				"nodeC": 1,
				"nodeD": 1,
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
				DecidedSfWitness: "a0_0",
				DecisiveRoots:    map[string]bool{"b2_2": true},
			},
			stakes{
				"nodeA": 1000000000000000000,
				"nodeB": 1,
				"nodeC": 1,
				"nodeD": 1,
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
				DecidedSfWitness: "a0_0",
				DecisiveRoots:    map[string]bool{"a4_4": true},
			},
			stakes{
				"nodeA": 4,
				"nodeB": 2,
				"nodeC": 1,
				"nodeD": 1,
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

	// members:
	var (
		mm         = make(members.Members, 0, len(peers))
	)
	for _, peer := range peers {
		mm.Add(peer, uint64(stakes[peer.String()]))
	}

	superMajority := get2of3(Amount(mm.TotalStake()))
	//t.Logf("superMajority = %s", superMajority.String())

	// events:
	events := make(map[hash.Event]*inter.Event)
	vertices := make(map[hash.Event]RootSlot)
	edges := make(map[fakeEdge]hash.Event)

	for dsc, root := range named {
		events[root.Hash()] = root
		h := root.Hash()

		vertices[h] = RootSlot{
			Frame: frameOf(dsc),
			Addr:  root.Creator,
		}
	}

	for dsc, root := range named {
		noPrev := false
		if strings.HasPrefix(dsc, "+") {
			noPrev = true
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

	election := NewElection(mm, Amount(mm.TotalStake()), superMajority, 0, stronglySeeFn)

	// ordering:
	var (
		processed      = make(map[hash.Event]*inter.Event)
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

func get2of3(x Amount) Amount {
	return x*2/3 + 1
}

func frameOf(dsc string) IdxFrame {
	s := strings.Split(dsc, "_")[1]
	h, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		panic(err)
	}
	return IdxFrame(h)
}
