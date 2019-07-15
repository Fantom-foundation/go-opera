package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// TODO: make Frame internal

// Frame is a consensus tables for frame.
type Frame struct {
	Index  idx.Frame
	Events EventsByPeer
	Roots  EventsByPeer

	save func()
}

// Save calls .save() if set.
func (f *Frame) Save() {
	if f.save != nil {
		f.save()
	}
}

// AddRoot appends root-event into frame.
func (f *Frame) AddRoot(e *Event) {
	if f.Roots.AddOne(e.Hash(), e.Creator) {
		f.Save()
	}
}

// AddEvent appends event into frame.
func (f *Frame) AddEvent(e *Event) {
	if f.Events.AddOne(e.Hash(), e.Creator) {
		f.Save()
	}
}

// ToWire converts to proto.Message.
func (f *Frame) ToWire() *wire.Frame {
	return &wire.Frame{
		Index:  uint32(f.Index),
		Events: f.Events.ToWire(),
		Roots:  f.Roots.ToWire(),
	}
}

// WireToFrame converts from wire.
func WireToFrame(w *wire.Frame) *Frame {
	if w == nil {
		return nil
	}
	return &Frame{
		Index:  idx.Frame(w.Index),
		Events: WireToEventsByPeer(w.Events),
		Roots:  WireToEventsByPeer(w.Roots),
	}
}

/*
 * Poset's methods:
 */

func (p *Poset) setFrameSaving(f *Frame) {
	f.save = func() {
		p.store.SetFrame(f, p.SuperFrameN)
	}
}
