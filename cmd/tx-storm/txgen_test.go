package main

import (
	"testing"
)

func TestGenerator(t *testing.T) {
	t.Skip("example only")
	g := newTxGenerator(0, 0, 20, 0)
	for i := 0; i < 2*len(g.accs); i++ {
		tx, info := g.Yield(99)
		t.Log(info.String(), tx.Raw.Nonce(), tx.Raw.Value())
	}
}

func TestCount(t *testing.T) {
	t.Skip("example only")
	for i := 0; i < 20; i++ {
		count := approximate(uint(i))
		t.Logf("%d ~= %d", i, count)
	}
}
