package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// TODO: make Frame internal

// Frame is a consensus tables for frame.
type Frame struct {
	Index     uint64
	FlagTable FlagTable
	NonRoots  Roots
	Balances  common.Hash

	save func()
}

// NodeRootsAdd appends root known for node.
func (f *Frame) NodeRootsAdd(node common.Address, roots Roots) {
	if f.FlagTable[node] == nil {
		f.FlagTable[node] = Roots{}
	}
	if f.FlagTable[node].Add(roots) {
		f.save()
	}
}

// NodeRootsGet returns roots of node. For read only, please.
func (f *Frame) NodeRootsGet(node common.Address) Roots {
	return f.FlagTable[node]
}

// NodeEventAdd appends event to frame.
func (f *Frame) NodeEventAdd(node common.Address, event EventHash) {
	if f.NonRoots[node] == nil {
		f.NonRoots[node] = EventHashes{}
	}
	if f.NonRoots[node].Add(event) {
		f.save()
	}
}

func (f *Frame) HasNodeEvent(node common.Address) bool {
	if f.Index == 0 {
		return false
	}
	if len(f.NonRoots[node]) > 0 {
		return true
	}
	if f.FlagTable[node] != nil && len(f.FlagTable[node][node]) > 0 {
		return true
	}
	return false
}

// SetBalances save PoS-balances state.
func (f *Frame) SetBalances(balances common.Hash) {
	if f.Balances != balances {
		f.Balances = balances
		f.save()
	}
}

/*
 * Poset's methods:
 */

func (p *Poset) saveFuncForFrame(f *Frame) func() {
	return func() {
		if f.Index > 0 {
			p.store.SetFrame(f)
		} else {
			panic("Frame 0 should be ephemeral")
		}
	}
}
