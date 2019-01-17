package common

import "hash/fnv"

// Hash64 TODO
func Hash64(data []byte) uint64 {
	h := fnv.New64a()

	_, _ = h.Write(data)

	return h.Sum64()
}
