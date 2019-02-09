package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// TODO: make Frame internal

// Frame
type Frame struct {
	Index    uint64
	Roots    EventHashes
	Balances common.Hash

	save func()
}

// IsRoot returns true if event is in roots list.
func (f *Frame) IsRoot(h EventHash) bool {
	if f.Index == 0 {
		return h.IsZero()
	}
	return f.Roots.Contains(h)
}

// SetRoot appends event to the roots list.
func (f *Frame) SetRoot(h EventHash) {
	f.Roots.Add(h)
	f.save()
}

// SetBalances save PoS-balanses state.
func (f *Frame) SetBalances(balances common.Hash) {
	f.Balances = balances
	f.save()
}

/*
 * Poset's methods:
 */

func (p *Poset) frame(n uint64) *Frame {
	if n < p.state.LastFinishedFrameN {
		panic("Too old frame requested")
	}
	// return ephemeral
	if n == 0 {
		return &Frame{
			Index:    0,
			Balances: p.state.Genesis,
		}
	}
	// return existing
	f := p.frames[n]
	if f != nil {
		return f
	}
	// create new frame
	f = &Frame{
		Index: n,
	}
	f.save = func() {
		if f.Index > 0 {
			p.store.SetFrame(f)
		} else {
			panic("Frame 0 should be ephemeral")
		}
	}
	p.frames[n] = f

	return f
}
