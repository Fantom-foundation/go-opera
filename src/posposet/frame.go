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
	Atroposes        timestampsByEvent
	Balances         common.Hash

	save func()
}

// Save calls .save() if set.
func (f *Frame) Save() {
	if f.save != nil {
		f.save()
	}
}

// AddRootsOf appends known roots for event.
func (f *Frame) AddRootsOf(event EventHash, roots eventsByNode) {
	if f.FlagTable[event] == nil {
		f.FlagTable[event] = eventsByNode{}
	}
	if f.FlagTable[event].Add(roots) {
		f.Save()
	}
}

// AddClothoCandidate adds event into ClothoCandidates list.
func (f *Frame) AddClothoCandidate(event EventHash, creator common.Address) {
	if f.ClothoCandidates.AddOne(event, creator) {
		f.Save()
	}
}

// SetAtropos makes Atropos from Clotho and consensus time.
func (f *Frame) SetAtropos(clotho EventHash, consensusTime Timestamp) {
	if t, ok := f.Atroposes[clotho]; ok && t == consensusTime {
		return
	}
	f.Atroposes[clotho] = consensusTime
	f.Save()
}

// GetRootsOf returns known roots of event. For read only, please.
func (f *Frame) GetRootsOf(event EventHash) eventsByNode {
	return f.FlagTable[event]
}

// SetBalances saves PoS-balances state.
func (f *Frame) SetBalances(balances common.Hash) bool {
	if f.Balances != balances {
		f.Balances = balances
		f.Save()
		return true
	}
	return false
}

/*
 * Poset's methods:
 */

func (p *Poset) setFrameSaving(f *Frame) {
	f.save = func() {
		if f.Index > p.state.LastFinishedFrameN {
			p.store.SetFrame(f)
		} else {
			panic(fmt.Errorf("Frame %d is finished and should not be changed", f.Index))
		}
	}
}
