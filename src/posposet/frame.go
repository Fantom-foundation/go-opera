package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// TODO: make Frame internal

// Frame
type Frame struct {
	Index    uint64
	Roots    map[EventHash]struct{}
	Balances common.Hash

	save func()
}

// IsRoot returns true if event is in roots list.
func (f *Frame) IsRoot(h EventHash) bool {
	if f.Index == 0 {
		return h.IsZero()
	}
	_, ok := f.Roots[h]
	return ok
}

// SetRoot appends event to the roots list.
func (f *Frame) SetRoot(h EventHash) {
	f.Roots[h] = struct{}{}
	f.save()
}

/*
 * Poset's methods:
 */

func (p *Poset) frame(n uint64) *Frame {
	if n == 0 {
		return &Frame{
			Index:    0,
			Balances: p.state.Genesis,
		}
	}
	f := p.store.GetFrame(n)
	if f == nil {
		f = p.newFrameFrom(p.frame(n - 1))
	}
	return f
}

func (p *Poset) newFrameFrom(prev *Frame) *Frame {
	f := &Frame{
		Index:    prev.Index + 1,
		Roots:    make(map[EventHash]struct{}),
		Balances: prev.Balances,
	}
	f.save = func() {
		if f.Index > 0 {
			p.store.SetFrame(f)
		} else {
			panic("Frame 0 should be ephemeral")
		}
	}

	return f
}
