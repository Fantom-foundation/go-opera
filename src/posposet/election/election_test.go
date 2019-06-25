package election

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

type fakeEdge struct {
	from RootHash
	to   RootSlot
}

type fakeRoot struct {
	Hash         RootHash
	Slot         RootSlot
	StronglySeen []RootHash
}

type processRootTest struct {
	Nodes         []ElectionNode
	SuperMajority *big.Int

	Roots []fakeRoot

	Answer *ElectionRes
}

// Allow short hashes for tests
func (h *RootHash) UnmarshalJSON(input []byte) error {
	if len(input) < 64 {
		var str string
		err := json.Unmarshal(input, &str)
		if err != nil {
			return err
		}
		*h = RootHash{common.HexToHash(str)}
	} else {
		return h.Hash.UnmarshalJSON(input)
	}
	return nil
}

// Allow short hashes for tests
func (h *NodeId) UnmarshalJSON(input []byte) error {
	if len(input) < 64 {
		var str string
		err := json.Unmarshal(input, &str)
		if err != nil {
			return err
		}
		*h = NodeId{common.HexToHash(str)}
	} else {
		return h.Hash.UnmarshalJSON(input)
	}
	return nil
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
				"stronglySeen" : ["a1", "b1", "c1", "d1"]
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
				"stronglySeen" : ["a1", "b1", "c1"]
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
				"stronglySeen" : ["a1", "b1", "c1", "d1"]
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
				"stronglySeen" : ["a3", "b3"]
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

		vertices := make(map[RootHash]RootSlot)
		edges := make(map[fakeEdge]RootHash)

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

		totalStake := new(big.Int)
		for _, node := range test.Nodes {
			totalStake = totalStake.Add(totalStake, node.StakeAmount)
		}

		election := NewElection(test.Nodes, totalStake, test.SuperMajority, 0, stronglySeeFn)

		for i, root := range test.Roots {
			decided, err := election.ProcessRoot(root.Hash, root.Slot)
			if err != nil {
				t.Fatal(name, err)
			}
			if i == len(test.Roots)-1 {
				// check refs
				if (test.Answer == nil) != (decided == nil) {
					t.Fatal(name, "expected ", test.Answer, "and calculated", decided)
				}
				// check values
				if (test.Answer != nil) && (*decided != *test.Answer) {
					t.Fatal(name, "expected ", test.Answer, "and calculated", decided)
				}
			} else if decided != nil {
				t.Fatal(name, "decision is made before last root in the test, on root", root.Hash.String())
			}
		}
	}
}
