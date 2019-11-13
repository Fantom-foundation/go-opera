package main

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestMakeAcc(t *testing.T) {
	const n = 10000

	exists := make(map[common.Address]uint, n)

	for i := uint(0); i < n; i++ {
		acc := MakeAcc(i)
		j, ok := exists[*acc.Addr]
		if ok {
			t.Fatalf("collision detected: %d and %d", i, j)
			return
		}

		exists[*acc.Addr] = i
	}
}
