package election

import (
	"encoding/json"
	"math/big"
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

type fakeEdge struct {
	from hash.Event
	to   RootSlot
}

type fakeRoot struct {
	Hash         hash.Event
	Slot         RootSlot
	StronglySeen []hash.Event
	Decisive     bool
}

type processRootTest struct {
	Nodes         []ElectionNode
	SuperMajority *big.Int

	Roots []fakeRoot

	Answer *ElectionRes
}

func TestProcessRootTest(t *testing.T) {
	testJson := `
{
    "TestProcessRoot_4_uni_notDecided" : {
		"nodes" : [
			{
				"nodeid" : "a",
				"stakeAmount" : 1
			},
			{
				"nodeid" : "b",
				"stakeAmount" : 1
			},
			{
				"nodeid" : "c",
				"stakeAmount" : 1
			},
			{
				"nodeid" : "d",
				"stakeAmount" : 1
			}
		],
		"superMajority" : 3,
		"roots" : [
			{
				"hash" : "a0",
				"slot" : {"frame" : 0, "nodeid" : "a"},
				"stronglySeen" : []
			},
			{
				"hash" : "b0",
				"slot" : {"frame" : 0, "nodeid" : "b"},
				"stronglySeen" : []
			},
			{
				"hash" : "c0",
				"slot" : {"frame" : 0, "nodeid" : "c"},
				"stronglySeen" : []
			},
			{
				"hash" : "d0",
				"slot" : {"frame" : 0, "nodeid" : "d"},
				"stronglySeen" : []
			},

			{
				"hash" : "a1",
				"slot" : {"frame" : 1, "nodeid" : "a"},
				"stronglySeen" : ["a0", "b0", "c0"]
			},
			{
				"hash" : "b1",
				"slot" : {"frame" : 1, "nodeid" : "b"},
				"stronglySeen" : ["a0", "b0", "c0"]
			},
			{
				"hash" : "c1",
				"slot" : {"frame" : 1, "nodeid" : "c"},
				"stronglySeen" : ["b0", "c0", "d0"]
			},
			{
				"hash" : "d1",
				"slot" : {"frame" : 1, "nodeid" : "d"},
				"stronglySeen" : ["b0", "c0", "d0"]
			},

			{
				"hash" : "a2",
				"slot" : {"frame" : 2, "nodeid" : "a"},
				"stronglySeen" : ["a1", "b1", "c1", "d1"]
			}
		],
		"answer" : null
	},
    "TestProcessRoot_4_uni_decided" : {
		"nodes" : [
			{
				"nodeid" : "a",
				"stakeAmount" : 1
			},
			{
				"nodeid" : "b",
				"stakeAmount" : 1
			},
			{
				"nodeid" : "c",
				"stakeAmount" : 1
			},
			{
				"nodeid" : "d",
				"stakeAmount" : 1
			}
		],
		"superMajority" : 3,
		"roots" : [
			{
				"hash" : "a0",
				"slot" : {"frame" : 0, "nodeid" : "a"},
				"stronglySeen" : []
			},
			{
				"hash" : "b0",
				"slot" : {"frame" : 0, "nodeid" : "b"},
				"stronglySeen" : []
			},
			{
				"hash" : "c0",
				"slot" : {"frame" : 0, "nodeid" : "c"},
				"stronglySeen" : []
			},
			{
				"hash" : "d0",
				"slot" : {"frame" : 0, "nodeid" : "d"},
				"stronglySeen" : []
			},

			{
				"hash" : "a1",
				"slot" : {"frame" : 1, "nodeid" : "a"},
				"stronglySeen" : ["a0", "b0", "c0"]
			},
			{
				"hash" : "b1",
				"slot" : {"frame" : 1, "nodeid" : "b"},
				"stronglySeen" : ["b0", "c0", "d0"]
			},
			{
				"hash" : "c1",
				"slot" : {"frame" : 1, "nodeid" : "c"},
				"stronglySeen" : ["b0", "c0", "d0"]
			},
			{
				"hash" : "d1",
				"slot" : {"frame" : 1, "nodeid" : "d"},
				"stronglySeen" : ["b0", "c0", "d0"]
			},

			{
				"hash" : "a2",
				"slot" : {"frame" : 2, "nodeid" : "a"},
				"stronglySeen" : ["a1", "b1", "c1", "d1"],
				"decisive" : true
			}
		],
		"answer" : {
			"decidedFrame" : 0,
			"decidedSfWitness" : "b0"
		}
	},
    "TestProcessRoot_4_uni_missingRoot_decided" : {
		"nodes" : [
			{
				"nodeid" : "a",
				"stakeAmount" : 1
			},
			{
				"nodeid" : "b",
				"stakeAmount" : 1
			},
			{
				"nodeid" : "c",
				"stakeAmount" : 1
			},
			{
				"nodeid" : "d",
				"stakeAmount" : 1
			}
		],
		"superMajority" : 3,
		"roots" : [
			{
				"hash" : "a0",
				"slot" : {"frame" : 0, "nodeid" : "a"},
				"stronglySeen" : []
			},
			{
				"hash" : "b0",
				"slot" : {"frame" : 0, "nodeid" : "b"},
				"stronglySeen" : []
			},
			{
				"hash" : "c0",
				"slot" : {"frame" : 0, "nodeid" : "c"},
				"stronglySeen" : []
			},

			{
				"hash" : "a1",
				"slot" : {"frame" : 1, "nodeid" : "a"},
				"stronglySeen" : ["a0", "b0", "c0"]
			},
			{
				"hash" : "b1",
				"slot" : {"frame" : 1, "nodeid" : "b"},
				"stronglySeen" : ["a0", "b0", "c0"]
			},
			{
				"hash" : "c1",
				"slot" : {"frame" : 1, "nodeid" : "c"},
				"stronglySeen" : ["a0", "b0", "c0"]
			},

			{
				"hash" : "a2",
				"slot" : {"frame" : 2, "nodeid" : "a"},
				"stronglySeen" : ["a1", "b1", "c1"],
				"decisive" : true
			}
		],
		"answer" : {
			"decidedFrame" : 0,
			"decidedSfWitness" : "a0"
		}
	},
    "TestProcessRoot_4_differentStakes_decided" : {
		"nodes" : [
			{
				"nodeid" : "a",
				"stakeAmount" : 100000000000000000000000000
			},
			{
				"nodeid" : "b",
				"stakeAmount" : 1
			},
			{
				"nodeid" : "c",
				"stakeAmount" : 1
			},
			{
				"nodeid" : "d",
				"stakeAmount" : 1
			}
		],
		"superMajority" : 70000000000000000000000000,
		"roots" : [
			{
				"hash" : "a0",
				"slot" : {"frame" : 0, "nodeid" : "a"},
				"stronglySeen" : []
			},
			{
				"hash" : "b0",
				"slot" : {"frame" : 0, "nodeid" : "b"},
				"stronglySeen" : []
			},
			{
				"hash" : "c0",
				"slot" : {"frame" : 0, "nodeid" : "c"},
				"stronglySeen" : []
			},
			{
				"hash" : "d0",
				"slot" : {"frame" : 0, "nodeid" : "d"},
				"stronglySeen" : []
			},

			{
				"hash" : "a1",
				"slot" : {"frame" : 1, "nodeid" : "a"},
				"stronglySeen" : ["a0", "b0", "c0"]
			},
			{
				"hash" : "b1",
				"slot" : {"frame" : 1, "nodeid" : "b"},
				"stronglySeen" : ["a0"]
			},
			{
				"hash" : "c1",
				"slot" : {"frame" : 1, "nodeid" : "c"},
				"stronglySeen" : ["a0"]
			},
			{
				"hash" : "d1",
				"slot" : {"frame" : 1, "nodeid" : "d"},
				"stronglySeen" : ["a0", "b0", "c0", "d0"]
			},

			{
				"hash" : "b2",
				"slot" : {"frame" : 2, "nodeid" : "b"},
				"stronglySeen" : ["a1", "b1", "c1", "d1"],
				"decisive" : true
			}
		],
		"answer" : {
			"decidedFrame" : 0,
			"decidedSfWitness" : "a0"
		}
	},
    "TestProcessRoot_4_differentStakes_5rounds_decided" : {
		"nodes" : [
			{
				"nodeid" : "a",
				"stakeAmount" : 4
			},
			{
				"nodeid" : "b",
				"stakeAmount" : 2
			},
			{
				"nodeid" : "c",
				"stakeAmount" : 1
			},
			{
				"nodeid" : "d",
				"stakeAmount" : 1
			}
		],
		"superMajority" : 6,
		"roots" : [
			{
				"hash" : "a0",
				"slot" : {"frame" : 0, "nodeid" : "a"},
				"stronglySeen" : []
			},
			{
				"hash" : "b0",
				"slot" : {"frame" : 0, "nodeid" : "b"},
				"stronglySeen" : []
			},
			{
				"hash" : "c0",
				"slot" : {"frame" : 0, "nodeid" : "c"},
				"stronglySeen" : []
			},
			{
				"hash" : "d0",
				"slot" : {"frame" : 0, "nodeid" : "d"},
				"stronglySeen" : []
			},

			{
				"hash" : "a1",
				"slot" : {"frame" : 1, "nodeid" : "a"},
				"stronglySeen" : ["a0", "b0"]
			},
			{
				"hash" : "b1",
				"slot" : {"frame" : 1, "nodeid" : "b"},
				"stronglySeen" : ["a0", "c0", "d0"]
			},
			{
				"hash" : "c1",
				"slot" : {"frame" : 1, "nodeid" : "c"},
				"stronglySeen" : ["a0", "c0", "d0"]
			},
			{
				"hash" : "d1",
				"slot" : {"frame" : 1, "nodeid" : "d"},
				"stronglySeen" : ["a0", "c0", "d0"]
			},

			{
				"hash" : "a2",
				"slot" : {"frame" : 2, "nodeid" : "a"},
				"stronglySeen" : ["a1", "b1"]
			},
			{
				"hash" : "b2",
				"slot" : {"frame" : 2, "nodeid" : "b"},
				"stronglySeen" : ["a1", "b1", "c1", "d1"]
			},
			{
				"hash" : "c2",
				"slot" : {"frame" : 2, "nodeid" : "c"},
				"stronglySeen" : ["a1", "b1", "c1", "d1"]
			},
			{
				"hash" : "d2",
				"slot" : {"frame" : 2, "nodeid" : "d"},
				"stronglySeen" : ["a1", "b1"]
			},

			{
				"hash" : "a3",
				"slot" : {"frame" : 3, "nodeid" : "a"},
				"stronglySeen" : ["a2", "b2", "c2", "d2"]
			},
			{
				"hash" : "b3",
				"slot" : {"frame" : 3, "nodeid" : "b"},
				"stronglySeen" : ["a2", "b2", "c2", "d2"]
			},
			{
				"hash" : "c3",
				"slot" : {"frame" : 3, "nodeid" : "c"},
				"stronglySeen" : ["a2", "b2", "c2", "d2"]
			},
			{
				"hash" : "d3",
				"slot" : {"frame" : 3, "nodeid" : "d"},
				"stronglySeen" : ["a2", "b2", "c2", "d2"]
			},

			{
				"hash" : "a4",
				"slot" : {"frame" : 4, "nodeid" : "a"},
				"stronglySeen" : ["a3", "b3"],
				"decisive" : true
			}
		],
		"answer" : {
			"decidedFrame" : 0,
			"decidedSfWitness" : "a0"
		}
	}
}
`
	tests := make(map[string]processRootTest)
	err := json.Unmarshal([]byte(testJson), &tests)
	if err != nil {
		t.Fatal(err)
	}

	for name, test := range tests {
		if len(test.Roots) == 0 {
			t.Fatal(name, "Empty test")
		}

		// define stronglySeeFn
		vertices := make(map[hash.Event]RootSlot)
		edges := make(map[fakeEdge]hash.Event)

		for _, root := range test.Roots {
			vertices[root.Hash] = root.Slot
		}

		for _, root := range test.Roots {
			for _, sSeen := range root.StronglySeen {
				edge := fakeEdge{
					from: root.Hash,
					to:   vertices[sSeen],
				}
				edges[edge] = sSeen
			}
		}

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

		totalStake := new(big.Int)
		for _, node := range test.Nodes {
			totalStake = totalStake.Add(totalStake, node.StakeAmount)
		}

		// run election
		election := NewElection(test.Nodes, totalStake, test.SuperMajority, 0, stronglySeeFn)

		ordered := fakeRoots(test.Roots).ByRand().BySeen()
		alreadyDecided := false
		for _, root := range ordered {
			decided, err := election.ProcessRoot(root.Hash, root.Slot)
			if err != nil {
				t.Fatal(name, err)
			}
			if root.Decisive || alreadyDecided {
				// check refs
				if (test.Answer == nil) != (decided == nil) {
					t.Fatal(name, "expected ", test.Answer, "and calculated", decided)
				}
				// check values
				if (test.Answer != nil) && (*decided != *test.Answer) {
					t.Fatal(name, "expected ", test.Answer, "and calculated", decided)
				}
				alreadyDecided = true
			} else if decided != nil {
				t.Fatal(name, "decision is made before last root in the test, on root", root.Hash.String())
			}
		}
	}
}

/*
 * root order:
 */

type fakeRoots []fakeRoot

// ByParents returns unordered roots.
func (ee fakeRoots) ByRand() (res fakeRoots) {
	res = make(fakeRoots, len(ee))

	mix := rand.Perm(len(ee))
	for i, j := range mix {
		res[i] = ee[j]
	}

	return
}

// ByParents returns roots ordered by seen dependency.
// It is a copy of inter.Events.ByParents().
func (ee fakeRoots) BySeen() (res fakeRoots) {
	unsorted := make(fakeRoots, len(ee))
	exists := hash.Events{}
	for i, e := range ee {
		unsorted[i] = e
		exists.Add(e.Hash)
	}
	ready := hash.Events{}
	for len(unsorted) > 0 {
	EVENTS:
		for i, e := range unsorted {

			for _, p := range e.StronglySeen {
				if exists.Contains(p) && !ready.Contains(p) {
					continue EVENTS
				}
			}

			res = append(res, e)
			unsorted = append(unsorted[0:i], unsorted[i+1:]...)
			ready.Add(e.Hash)
			break
		}
	}

	return
}
