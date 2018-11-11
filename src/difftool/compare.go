package difftool

import (
	"fmt"
	"reflect"

	"github.com/andrecronje/lachesis/src/node"
	"github.com/andrecronje/lachesis/src/poset"
)

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

func compare(n0, n1 *node.Node) (diff *Diff) {
	diff = &Diff{
		node: [2]*node.Node{n0, n1},
		IDs:  [2]int{n0.ID(), n1.ID()},
	}

	compareBlocks(diff)

	compareConsensus(diff)

	return
}

func compareBlocks(diff *Diff) {
	var n0, n1 = diff.node[0], diff.node[1]
	var minH, tmp int
	if minH, tmp = n0.GetLastBlockIndex(), n1.GetLastBlockIndex(); minH > tmp {
		minH, tmp = tmp, minH
	}
	if minH != tmp {
		diff.BlocksGap = tmp - minH
	}

	var b0, b1 poset.Block
	for i := 0; i <= minH; i++ {
		b0, diff.Err = n0.GetBlock(i)
		if diff.Err != nil {
			return
		}
		b1, diff.Err = n1.GetBlock(i)
		if diff.Err != nil {
			return
		}

		// NOTE: the same blocks Hashes are different because their Signatures.
		// So, compare bodies only.
		if !reflect.DeepEqual(b0.Body, b1.Body) {
			diff.FirstBlockIndex = i
			diff.Descr = fmt.Sprintf("%+v \n!= \n%+v\n", b0.Body, b1.Body)
			return
		}
	}
}

func compareConsensus(diff *Diff) {
	var n0, n1 = diff.node[0], diff.node[1]

	if r0, r1 := n0.GetLastRound(), n1.GetLastRound(); r0 != r1 {
		if r0 < r1 {
			r0, r1 = r1, r0
		}
		diff.RoundGap = r0 - r1
	}
}
