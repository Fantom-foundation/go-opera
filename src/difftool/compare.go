package difftool

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/andrecronje/lachesis/src/node"
	"github.com/andrecronje/lachesis/src/poset"
)

// Compare compares each node with others
func Compare(nodes ...*node.Node) (result Result) {
	for i := len(nodes) - 1; i > 0; i-- {
		n0 := nodes[i]
		for _, n1 := range nodes[:i] {
			diff := compare(n0, n1)
			result = append(result, diff)
		}
	}
	return
}

// compare compares pair of nodes
func compare(n0, n1 *node.Node) (diff *Diff) {
	diff = &Diff{
		node: [2]*node.Node{n0, n1},
		IDs:  [2]int64{n0.ID(), n1.ID()},
	}

	if !compareBlocks(diff) {
		return
	}
	if !compareRounds(diff) {
		return
	}
	if !compareFrames(diff) {
		return
	}

	return
}

// compareBlocks returns true if we need to go deeper
func compareBlocks(diff *Diff) bool {
	var n0, n1 = diff.node[0], diff.node[1]

	minH, tmp := n0.GetLastBlockIndex(), n1.GetLastBlockIndex()
	diff.BlocksGap = minH - tmp
	if minH > tmp {
		minH, tmp = tmp, minH
	}

	var b0, b1 poset.Block
	var i int64
	for i = 0; i <= minH; i++ {
		b0, diff.Err = n0.GetBlock(i)
		if diff.Err != nil {
			return false
		}
		b1, diff.Err = n1.GetBlock(i)
		if diff.Err != nil {
			return false
		}

		// NOTE: the same blocks Hashes are different because their Signatures.
		// So, compare bodies only.
		if !reflect.DeepEqual(b0.Body, b1.Body) {
			diff.FirstBlockIndex = i
			diff.AddDescr(fmt.Sprintf("block:\n%+v \n!= \n%+v\n", b0.Body, b1.Body))

			diff.FirstRoundIndex = b0.RoundReceived()
			if diff.FirstRoundIndex > b1.RoundReceived() {
				diff.FirstRoundIndex = b1.RoundReceived()
			}

			return true
		}
	}

	return false
}

// compareRounds returns true if we need to go deeper
func compareRounds(diff *Diff) bool {
	var n0, n1 = diff.node[0], diff.node[1]

	diff.RoundGap = n0.GetLastRound() - n1.GetLastRound()

	var r0, r1 poset.RoundInfo
	var i int64
	for i = 0; i <= diff.FirstRoundIndex; i++ {

		r0, diff.Err = n0.GetRound(i)
		if diff.Err != nil {
			return false
		}
		r1, diff.Err = n1.GetRound(i)
		if diff.Err != nil {
			return false
		}

		if !reflect.DeepEqual(r0, r1) {
			diff.FirstRoundIndex = i
			diff.AddDescr(fmt.Sprintf("round:\n%+v \n!= \n%+v\n", r0, r1))
			return true
		}

		w0, w1 := n0.RoundWitnesses(i), n1.RoundWitnesses(i)
		sort.Sort(ByValue(w0))
		sort.Sort(ByValue(w1))
		if !reflect.DeepEqual(w0, w1) {
			diff.FirstRoundIndex = i
			diff.AddDescr(fmt.Sprintf("witness:\n%+v \n!= \n%+v\n", w0, w1))
			return true
		}
	}

	return false
}

// compareFrames returns true if we need to go deeper
func compareFrames(diff *Diff) bool {
	var n0, n1 = diff.node[0], diff.node[1]

	var f0, f1 poset.Frame
	f0, diff.Err = n0.GetFrame(diff.FirstRoundIndex)
	if diff.Err != nil {
		return false
	}
	f1, diff.Err = n1.GetFrame(diff.FirstRoundIndex)
	if diff.Err != nil {
		return false
	}

	if !reflect.DeepEqual(f0, f1) {
		diff.AddDescr(fmt.Sprintf("frame:\n%+v \n!= \n%+v\n", f0, f1))
		return true
	}

	return false
}
