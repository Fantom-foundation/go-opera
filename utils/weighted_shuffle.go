package utils

import (
	"crypto/sha256"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/common/littleendian"
)

type weightedShuffleNode struct {
	thisWeight  pos.Stake
	leftWeight  pos.Stake
	rightWeight pos.Stake
}

type weightedShuffleTree struct {
	seed      common.Hash
	seedIndex int

	weights []pos.Stake
	nodes   []weightedShuffleNode
}

func (t *weightedShuffleTree) leftIndex(i int) int {
	return i*2 + 1
}

func (t *weightedShuffleTree) rightIndex(i int) int {
	return i*2 + 2
}

func (t *weightedShuffleTree) build(i int) pos.Stake {
	if i >= len(t.weights) {
		return 0
	}
	this_w := t.weights[i]
	left_w := t.build(t.leftIndex(i))
	right_w := t.build(t.rightIndex(i))

	if this_w <= 0 {
		panic("all the weight must be positive")
	}

	t.nodes[i] = weightedShuffleNode{
		thisWeight:  this_w,
		leftWeight:  left_w,
		rightWeight: right_w,
	}
	return this_w + left_w + right_w
}

func (t *weightedShuffleTree) rand64() uint64 {
	if t.seedIndex == 32 {
		hasher := sha256.New() // use sha2 instead of sha3 for speed
		hasher.Write(t.seed.Bytes())
		t.seed = common.BytesToHash(hasher.Sum(nil))
		t.seedIndex = 0
	}
	// use not used parts of old seed, instead of calculating new one
	res := littleendian.BytesToInt64(t.seed[t.seedIndex : t.seedIndex+8])
	t.seedIndex += 8
	return res
}

func (t *weightedShuffleTree) retrieve(i int) int {
	node := t.nodes[i]
	total := node.rightWeight + node.leftWeight + node.thisWeight

	r := pos.Stake(t.rand64()) % total

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

// Builds weighted random permutation
// Returns first {size} entries of {weights} permutation.
// Call with {size} == len(weights) to get the whole permutation.
func WeightedPermutation(size int, weights []pos.Stake, seed common.Hash) []int {
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
