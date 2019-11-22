package election

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/utils"
)

type fakeEdge struct {
	from hash.Event
	to   hash.Event
}

type (
	stakes map[string]pos.Stake
)

type testExpected struct {
	DecidedFrame   idx.Frame
	DecidedAtropos string
	DecisiveRoots  map[string]bool
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
				DecidedFrame:   0,
				DecidedAtropos: "b0_0",
				DecisiveRoots:  map[string]bool{"a2_2": true},
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
				DecidedFrame:   0,
				DecidedAtropos: "b0_0",
				DecisiveRoots:  map[string]bool{"a2_2": true},
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
				DecidedFrame:   0,
				DecidedAtropos: "a0_0",
				DecisiveRoots:  map[string]bool{"b2_2": true},
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
				DecidedFrame:   0,
				DecidedAtropos: "a0_0",
				DecisiveRoots:  map[string]bool{"a4_4": true},
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

	// events:
	ordered := make(inter.Events, 0)
	events := make(map[hash.Event]*inter.Event)
	frameRoots := make(map[idx.Frame][]RootAndSlot)
	vertices := make(map[hash.Event]Slot)
	edges := make(map[fakeEdge]bool)

	peers, _, _ := inter.ASCIIschemeForEach(dag, inter.ForEachEvent{
		Process: func(root *inter.Event, name string) {
			// store all the events
			ordered = append(ordered, root)

			events[root.Hash()] = root

			slot := Slot{
				Frame: frameOf(name),
				Addr:  root.Creator,
			}
			vertices[root.Hash()] = slot

			frameRoots[frameOf(name)] = append(frameRoots[frameOf(name)], RootAndSlot{
				ID:   root.Hash(),
				Slot: slot,
			})

			// build edges to be able to fake forkless cause fn
			noPrev := false
			if strings.HasPrefix(name, "+") {
				noPrev = true
			}
			from := root.Hash()
			for _, observed := range root.Parents {
				if root.IsSelfParent(observed) && noPrev {
					continue
				}
				to := observed
				edge := fakeEdge{
					from: from,
					to:   to,
				}
				edges[edge] = true
			}
		},
	})

	// validators:
	var (
		vv = *pos.NewValidators()
	)
	for _, peer := range peers {
		vv.Set(peer, stakes[utils.NameOf(peer)])
	}

	// forkless cause func:
	forklessCauseFn := func(a hash.Event, b hash.Event) bool {
		edge := fakeEdge{
			from: a,
			to:   b,
		}
		return edges[edge]
	}
	getFrameRootsFn := func(f idx.Frame) []RootAndSlot {
		return frameRoots[f]
	}

	// re-order events randomly, preserving parents order
	unordered := make(inter.Events, len(ordered))
	for i, j := range rand.Perm(len(ordered)) {
		unordered[i] = ordered[j]
	}
	ordered = unordered.ByParents()

	election := New(vv, 0, forklessCauseFn, getFrameRootsFn)

	// processing:
	var alreadyDecided bool
	for _, root := range ordered {
		rootHash := root.Hash()
		rootSlot, ok := vertices[rootHash]
		if !ok {
			t.Fatal("inconsistent vertices")
		}
		got, err := election.ProcessRoot(RootAndSlot{
			ID:   rootHash,
			Slot: rootSlot,
		})
		if err != nil {
			t.Fatal(err)
		}

		// checking:
		decisive := expected != nil && expected.DecisiveRoots[root.Hash().String()]
		if decisive || alreadyDecided {
			assertar.NotNil(got)
			assertar.Equal(expected.DecidedFrame, got.Frame)
			assertar.Equal(expected.DecidedAtropos, got.Atropos.String())
			alreadyDecided = true
		} else {
			assertar.Nil(got)
		}
	}
}

func frameOf(dsc string) idx.Frame {
	s := strings.Split(dsc, "_")[1]
	h, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		panic(err)
	}
	return idx.Frame(h)
}
