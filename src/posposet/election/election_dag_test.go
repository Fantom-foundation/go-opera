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

type (
	stakes map[string]*big.Int
)

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
a1_1  ╬ ─ ─ ╣     ║
║     ║     ║     ║
║╚ ─  b1_1  ╣     ║
║     ║     ║     ║
║     ║╚ ─  c1_1  ╣
║     ║     ║     ║
║     ║╚  ─ ╫╩  ─ d1_1
║     ║     ║     ║
a2_2  ╬ ─ ─ ╬ ─ ─ ╣
║     ║     ║     ║
`)
	})

	t.Run("4 uni decided", func(t *testing.T) {
		testProcessRoot(t,
			&ElectionRes{
				DecidedFrame:     0,
				DecidedSfWitness: "d0_0", // NOTE: but b0_0 in original
			},
			stakes{
				"nodeA": big.NewInt(1),
				"nodeB": big.NewInt(1),
				"nodeC": big.NewInt(1),
				"nodeD": big.NewInt(1),
			}, `
a0_0  b0_0  c0_0  d0_0
║     ║     ║     ║
a1_1  ╬ ─ ─ ╣     ║
║     ║     ║     ║
║     b1_1  ╬ ─ ─ ╣
║     ║     ║     ║
║     ║╚ ─  c1_1  ╣
║     ║     ║     ║
║     ║╚  ─ ╫╩  ─ d1_1
║     ║     ║     ║
a2_2  ╬ ─ ─ ╬ ─ ─ ╣
║     ║     ║     ║
`)
	})

	t.Run("4 uni missingRoot decided", func(t *testing.T) {
		testProcessRoot(t,
			&ElectionRes{
				DecidedFrame:     0,
				DecidedSfWitness: "c0_0", // NOTE: but a0_0 in original
			},
			stakes{
				"nodeA": big.NewInt(1),
				"nodeB": big.NewInt(1),
				"nodeC": big.NewInt(1),
				"nodeD": big.NewInt(1),
			}, `
a0_0  b0_0  c0_0  d0_0
║     ║     ║     ║
a1_1  ╬ ─ ─ ╣     ║
║     ║     ║     ║
║╚ ─  b1_1  ╣     ║
║     ║     ║     ║
║╚  ─ ╫╩ ─  c1_1  ║
║     ║     ║     ║
a2_2  ╬ ─ ─ ╣     ║
║     ║     ║     ║
`)
	})

	t.Run("4 differentStakes decided", func(t *testing.T) {
		testProcessRoot(t,
			&ElectionRes{
				DecidedFrame:     0,
				DecidedSfWitness: "b0_0", // NOTE: but a0_0 in original
			},
			stakes{
				"nodeA": big.NewInt(1000000000000000000),
				"nodeB": big.NewInt(1),
				"nodeC": big.NewInt(1),
				"nodeD": big.NewInt(1),
			}, `
a0_0  b0_0  c0_0  d0_0
║     ║     ║     ║
a1_1  ╬ ─ ─ ╣     ║
║     ║     ║     ║
║╚  ─ b1_1  ║     ║
║     ║     ║     ║
║╚  ─ ╫ ─ ─ c1_1  ║
║     ║     ║     ║
║╚  ─ ╫╩  ─ ╫╩  ─ d1_1
║     ║     ║     ║
╠ ─ ─ b2_2  ╬ ─ ─ ╣
║     ║     ║     ║
`)
	})

	t.Run("4 differentStakes 5rounds decided", func(t *testing.T) {
		testProcessRoot(t,
			&ElectionRes{
				DecidedFrame:     0,
				DecidedSfWitness: "a0_0",
			},
			stakes{
				"nodeA": big.NewInt(4),
				"nodeB": big.NewInt(2),
				"nodeC": big.NewInt(1),
				"nodeD": big.NewInt(1),
			}, `
a0_0  b0_0  c0_0  d0_0
║     ║     ║     ║
a1_1  ╣     ║     ║
║     ║     ║     ║
║     b1_1  ╬ ─ ─ ╣
║     ║     ║     ║
║╚  ─ ╫ ─ ─ c1_1  ╣
║     ║     ║     ║
║╚  ─ ╫ ─ ─ ╫╩  ─ d1_1
║     ║     ║     ║
a2_2  ╣     ║     ║
║     ║     ║     ║
║╚  ─ b2_2  ╬ ─ ─ ╣
║     ║     ║     ║
║╚  ─ ╫╩  ─ c2_2  ╣
║     ║     ║     ║
║╚  ─ ╫╩ ─  ╫ ─ ─ d2_2
║     ║     ║     ║
a3_3  ╬ ─ ─ ╬ ─ ─ ╣
║     ║     ║     ║
║╚  ─ b3_3  ╬ ─ ─ ╣
║     ║     ║     ║
║╚  ─ ╫╩  ─ c3_3  ╣
║     ║     ║     ║
║╚  ─ ╫╩ ─  ╫╩  ─ d3_3
║     ║     ║     ║
a4_4  ╣     ║     ║
║     ║     ║     ║
`)
	})

}

func testProcessRoot(
	t *testing.T,
	expected *ElectionRes,
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
			Nodeid:      NodeId{peer},
			StakeAmount: stakes[peer.String()],
		}
		nodes = append(nodes, n)
		totalStake.Add(totalStake, n.StakeAmount)
	}

	superMajority := get2of3(totalStake)
	//t.Logf("superMajority = %s", superMajority.String())

	// events:
	vertices := make(map[RootHash]RootSlot)
	edges := make(map[fakeEdge]RootHash)

	for dsc, root := range named {
		h := RootHash{root.Hash()}

		vertices[h] = RootSlot{
			Frame:  frameOf(dsc),
			Nodeid: NodeId{root.Creator},
		}
	}

	for _, root := range named {
		for sSeen := range root.Parents {
			from := RootHash{root.Hash()}
			to := RootHash{sSeen}

			edge := fakeEdge{
				from: from,
				to:   vertices[to],
			}
			edges[edge] = to
		}
	}

	// election:
	stronglySeeFn := func(a RootHash, b RootSlot) *RootHash {
		edge := fakeEdge{
			from: a,
			to: RootSlot{
				Frame:  b.Frame,
				Nodeid: b.Nodeid,
			},
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
		err       error
		steps     = 0
		processed = make(map[hash.Event]*inter.Event)
		got       *ElectionRes
	)
	orderThenProcess := ordering.EventBuffer(
		// process
		func(root *inter.Event) {
			if got != nil {
				return
			}
			steps++

			rootHash := RootHash{root.Hash()}
			rootSlot, ok := vertices[rootHash]
			if !ok {
				t.Fatal("inconsistent vertices")
			}
			got, err = election.ProcessRoot(rootHash, rootSlot)
			if err != nil {
				t.Fatal(err)
			}
			processed[root.Hash()] = root
		},
		// drop
		func(e *inter.Event, err error) {
			t.Fatal(e, err)
		},
		// exists
		func(h hash.Event) *inter.Event {
			return processed[h]
		},
	)

	// processing:
	for _, root := range named {
		orderThenProcess(root)
		if got != nil {
			break
		}
	}

	// checking:
	// assertar.Equal(len(named), steps, "decision is made before last root") // NOTE: is not stable
	assertar.Equal(expected, got)
}

func get2of3(x *big.Int) *big.Int {
	res := new(big.Int)
	mod := new(big.Int)
	res.
		Mul(x, big.NewInt(2)).
		DivMod(res, big.NewInt(3), mod)

	if mod.Cmp(big.NewInt(0)) > 0 {
		res.Add(res, big.NewInt(1))
	}

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
