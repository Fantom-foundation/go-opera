package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// Frame
type Frame struct {
	Index    uint64
	Roots    map[EventHash]struct{}
	Balances common.Hash

	save func()
}

func StartNewFrame(prev uint64, save func(*Frame)) *Frame {
	f := &Frame{
		Index:    prev + 1,
		Roots:    make(map[EventHash]struct{}),
		Balances: common.Hash{}, // TODO: replace with genesis hash
	}
	f.save = func() {
		if f.Index > 0 {
			save(f)
		} else {
			panic("Frame 0 should be ephemeral")
		}
	}

	return f
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
