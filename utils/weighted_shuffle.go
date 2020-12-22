package utils

import (
	"crypto/sha256"

	"github.com/Fantom-foundation/lachesis-base/common/littleendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
)

type weightedShuffleNode struct {
	thisWeight  pos.Weight
	leftWeight  pos.Weight
	rightWeight pos.Weight
}

type weightedShuffleTree struct {
	seed      hash.Hash
	seedIndex int

	weights []pos.Weight
	nodes   []weightedShuffleNode
}

func (t *weightedShuffleTree) leftIndex(i int) int {
	return i*2 + 1
}

func (t *weightedShuffleTree) rightIndex(i int) int {
	return i*2 + 2
}

func (t *weightedShuffleTree) build(i int) pos.Weight {
	if i >= len(t.weights) {
		return 0
	}
	thisW := t.weights[i]
	leftW := t.build(t.leftIndex(i))
	rightW := t.build(t.rightIndex(i))

	if thisW <= 0 {
		panic("all the weight must be positive")
	}

	t.nodes[i] = weightedShuffleNode{
		thisWeight:  thisW,
		leftWeight:  leftW,
		rightWeight: rightW,
	}
	return thisW + leftW + rightW
}

func (t *weightedShuffleTree) rand32() uint32 {
	if t.seedIndex == 32 {
		hasher := sha256.New() // use sha2 instead of sha3 for speed
		hasher.Write(t.seed.Bytes())
		t.seed = hash.BytesToHash(hasher.Sum(nil))
		t.seedIndex = 0
	}
	// use not used parts of old seed, instead of calculating new one
	res := littleendian.BytesToUint32(t.seed[t.seedIndex : t.seedIndex+4])
	t.seedIndex += 4
	return res
}

func (t *weightedShuffleTree) retrieve(i int) int {
	node := t.nodes[i]
	total := node.rightWeight + node.leftWeight + node.thisWeight

	r := pos.Weight(t.rand32()) % total

	if r < node.thisWeight {
		t.nodes[i].thisWeight = 0
		return i
	} else if r < node.thisWeight+node.leftWeight {
		chosen := t.retrieve(t.leftIndex(i))
		t.nodes[i].leftWeight -= t.weights[chosen]
		return chosen
	} else {
		chosen := t.retrieve(t.rightIndex(i))
		t.nodes[i].rightWeight -= t.weights[chosen]
		return chosen
	}
}

// WeightedPermutation builds weighted random permutation
// Returns first {size} entries of {weights} permutation.
// Call with {size} == len(weights) to get the whole permutation.
func WeightedPermutation(size int, weights []pos.Weight, seed hash.Hash) []int {
	if len(weights) < size {
		panic("the permutation size must be less or equal to weights size")
	}

	if len(weights) == 0 {
		return make([]int, 0)
	}

	tree := weightedShuffleTree{
		weights: weights,
		nodes:   make([]weightedShuffleNode, len(weights)),
		seed:    seed,
	}
	tree.build(0)

	permutation := make([]int, size)
	for i := 0; i < size; i++ {
		permutation[i] = tree.retrieve(0)
	}
	return permutation
}
