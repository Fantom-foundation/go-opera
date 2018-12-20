package common

import "hash/fnv"

// Hash32 TODO
func Hash32(data []byte) int {
	h := fnv.New32a()

	h.Write(data)

	return int(h.Sum32())
}
