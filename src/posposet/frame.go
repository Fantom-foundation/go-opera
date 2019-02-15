package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// TODO: make Frame internal

// Frame is a consensus tables for frame.
type Frame struct {
	Index     uint64
	FlagTable FlagTable
	Balances  common.Hash

	save func()
}

// EventRootsAdd appends known roots for event.
func (f *Frame) EventRootsAdd(event EventHash, roots Events) {
	if f.FlagTable[event] == nil {
		f.FlagTable[event] = Events{}
	}
	if f.FlagTable[event].Add(roots) {
		f.save()
	}
}

// EventRootsGet returns known roots of event. For read only, please.
func (f *Frame) EventRootsGet(event EventHash) Events {
	return f.FlagTable[event]
}

// HasNodeEvent returns true if event is in.
func (f *Frame) HasNodeEvent(node common.Address, event EventHash) bool {
	if f.Index == 0 {
		return false
	}
	if _, ok := f.FlagTable[event]; ok {
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
