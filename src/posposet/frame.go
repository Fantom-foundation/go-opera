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

// AddRootsOf appends known roots for event.
func (f *Frame) AddRootsOf(event EventHash, roots eventsByNode) {
	if f.FlagTable[event] == nil {
		f.FlagTable[event] = eventsByNode{}
	}
	if f.FlagTable[event].Add(roots) {
		f.save()
	}
}

// GetRootsOf returns known roots of event. For read only, please.
func (f *Frame) GetRootsOf(event EventHash) eventsByNode {
	return f.FlagTable[event]
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
