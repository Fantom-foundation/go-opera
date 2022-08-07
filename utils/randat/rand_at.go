package randat

import (
	"math/rand"
)

type cached struct {
	seed uint64
	r    uint64
}

var (
	gSeed = rand.Int63()
	cache = cached{}
)

// RandAt returns random number with seed
// Not safe for concurrent use
func RandAt(seed uint64) uint64 {
	if seed != 0 && cache.seed == seed {
		return cache.r
	}
	cache.seed = seed
	cache.r = rand.New(rand.NewSource(gSeed ^ int64(seed))).Uint64()
	return cache.r
}
