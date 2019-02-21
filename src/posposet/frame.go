package posposet

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// TODO: make Frame internal

// Frame is a consensus tables for frame.
type Frame struct {
	Index            uint64
	FlagTable        FlagTable
	ClothoCandidates eventsByNode
	ClothoList       eventsByNode
	Balances         common.Hash

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

// AddClothoCandidate adds event into ClothoCandidates list.
func (f *Frame) AddClothoCandidate(event EventHash, creator common.Address) {
	if f.ClothoCandidates.AddOne(event, creator) {
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
		if f.Index > p.state.LastFinishedFrameN {
			p.store.SetFrame(f)
		} else {
			panic(fmt.Errorf("Frame %d is finished and should not be changed", f.Index))
		}
	}
}
