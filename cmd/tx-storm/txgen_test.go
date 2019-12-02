package main

import (
	"testing"
)

func TestGenerator(t *testing.T) {
	t.Skip("example only")
	g := newTxGenerator(0, 20, 0)
	for i := 0; i < 2*len(g.accs); i++ {
		tx := g.Yield()
		t.Log(tx.Info.String(), tx.Raw.Nonce(), tx.Raw.Value())
	}
}
