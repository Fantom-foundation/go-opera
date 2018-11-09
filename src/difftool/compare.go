package difftool

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/andrecronje/lachesis/src/node"
	"github.com/andrecronje/lachesis/src/poset"
)

// Diff contains and prints differences details
type Diff struct {
	Err error `json:"-"`

	node            [2]*node.Node `json:"-"`
	IDs             [2]int
	BlocksGap       int
	FirstBlockIndex int

	Descr string `json:"-"`
}

func (d *Diff) IsEmpty() bool {
	// TODO: remove const true, it's for develop only
	has := d.FirstBlockIndex > 0 || true
	return !has
}

func (d *Diff) ToString() string {
	if d.Err != nil {
		return fmt.Sprintf("ERR: %s", d.Err.Error())
	}
	if d.IsEmpty() {
		return ""
	}

	raw, err := json.Marshal(d)
	if err != nil {
		return fmt.Sprintf("JSON: %s", err.Error())
	}
	return string(raw)
}

/*
 * Comparing
 */

func Compare(nodes ...*node.Node) (result []*Diff) {
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

	return
}

func compareBlocks(diff *Diff) {
	var minH, tmp int
	var n0, n1 = diff.node[0], diff.node[1]
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
