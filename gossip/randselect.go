package gossip

import (
	"math/rand"
)

// wrsItem interface should be implemented by any entries that are to be selected from
// a weightedRandomSelect set. Note that recalculating monotonously decreasing item
// weights on-demand (without constantly calling update) is allowed
type wrsItem interface {
	Weight() int64
}

// weightedRandomSelect is capable of weighted random selection from a set of items
type weightedRandomSelect struct {
	root *wrsNode
	idx  map[wrsItem]int
}

// newWeightedRandomSelect returns a new weightedRandomSelect structure
func newWeightedRandomSelect() *weightedRandomSelect {
	return &weightedRandomSelect{root: &wrsNode{maxItems: wrsBranches}, idx: make(map[wrsItem]int)}
}

// update updates an item's weight, adds it if it was non-existent or removes it if
// the new weight is zero. Note that explicitly updating decreasing weights is not necessary.
func (w *weightedRandomSelect) update(item wrsItem) {
	w.setWeight(item, item.Weight())
}

// remove removes an item from the set
func (w *weightedRandomSelect) remove(item wrsItem) {
	w.setWeight(item, 0)
}

// setWeight sets an item's weight to a specific value (removes it if zero)
func (w *weightedRandomSelect) setWeight(item wrsItem, weight int64) {
	idx, ok := w.idx[item]
	if ok {
		w.root.setWeight(idx, weight)
		if weight == 0 {
			delete(w.idx, item)
		}
	} else {
		if weight != 0 {
			if w.root.itemCnt == w.root.maxItems {
				// add a new level
				newRoot := &wrsNode{sumWeight: w.root.sumWeight, itemCnt: w.root.itemCnt, level: w.root.level + 1, maxItems: w.root.maxItems * wrsBranches}
				newRoot.items[0] = w.root
				newRoot.weights[0] = w.root.sumWeight
				w.root = newRoot
			}
			w.idx[item] = w.root.insert(item, weight)
		}
	}
}

// choose randomly selects an item from the set, with a chance proportional to its
// current weight. If the weight of the chosen element has been decreased since the
// last stored value, returns it with a newWeight/oldWeight chance, otherwise just
// updates its weight and selects another one
func (w *weightedRandomSelect) choose() wrsItem {
	for {
		if w.root.sumWeight == 0 {
			return nil
		}
		val := rand.Int63n(w.root.sumWeight)
		choice, lastWeight := w.root.choose(val)
		weight := choice.Weight()
		if weight != lastWeight {
			w.setWeight(choice, weight)
		}
		if weight >= lastWeight || rand.Int63n(lastWeight) < weight {
			return choice
		}
	}
}

const wrsBranches = 8 // max number of branches in the wrsNode tree

// wrsNode is a node of a tree structure that can store wrsItems or further wrsNodes.
type wrsNode struct {
	items                    [wrsBranches]interface{}
	weights                  [wrsBranches]int64
	sumWeight                int64
	level, itemCnt, maxItems int
}

// insert recursively inserts a new item to the tree and returns the item index
func (n *wrsNode) insert(item wrsItem, weight int64) int {
	branch := 0
	for n.items[branch] != nil && (n.level == 0 || n.items[branch].(*wrsNode).itemCnt == n.items[branch].(*wrsNode).maxItems) {
		branch++
		if branch == wrsBranches {
			panic(nil)
		}
	}
	n.itemCnt++
	n.sumWeight += weight
	n.weights[branch] += weight
	if n.level == 0 {
		n.items[branch] = item
		return branch
	}
	var subNode *wrsNode
	if n.items[branch] == nil {
		subNode = &wrsNode{maxItems: n.maxItems / wrsBranches, level: n.level - 1}
		n.items[branch] = subNode
	} else {
		subNode = n.items[branch].(*wrsNode)
	}
	subIdx := subNode.insert(item, weight)
	return subNode.maxItems*branch + subIdx
}

// setWeight updates the weight of a certain item (which should exist) and returns
// the change of the last weight value stored in the tree
func (n *wrsNode) setWeight(idx int, weight int64) int64 {
	if n.level == 0 {
		oldWeight := n.weights[idx]
		n.weights[idx] = weight
		diff := weight - oldWeight
		n.sumWeight += diff
		if weight == 0 {
			n.items[idx] = nil
			n.itemCnt--
		}
		return diff
	}
	branchItems := n.maxItems / wrsBranches
	branch := idx / branchItems
	diff := n.items[branch].(*wrsNode).setWeight(idx-branch*branchItems, weight)
	n.weights[branch] += diff
	n.sumWeight += diff
	if weight == 0 {
		n.itemCnt--
	}
	return diff
}

// choose recursively selects an item from the tree and returns it along with its weight
func (n *wrsNode) choose(val int64) (wrsItem, int64) {
	for i, w := range n.weights {
		if val < w {
			if n.level == 0 {
				return n.items[i].(wrsItem), n.weights[i]
			}
			return n.items[i].(*wrsNode).choose(val)
		}
		val -= w
	}
	panic(nil)
}
