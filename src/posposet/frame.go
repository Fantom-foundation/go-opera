package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// TODO: make Frame internal

// Frame is a consensus tables for frame.
type Frame struct {
	Index  idx.Frame
	Events EventsByPeer // TODO erase this index
	Roots  EventsByPeer

	timeOffset inter.Timestamp
	timeRatio  inter.Timestamp

	save func()
}

// Save calls .save() if set.
func (f *Frame) Save() {
	if f.save != nil {
		f.save()
	}
}

// GetConsensusTimestamp calc consensus timestamp for given event.
func (f *Frame) GetConsensusTimestamp(e *Event) inter.Timestamp {
	return inter.Timestamp(e.Lamport)*f.timeRatio + f.timeOffset
}

// SetTimes set new timeOffset and new TimeRatio.
func (f *Frame) SetTimes(offset, ratio inter.Timestamp) {
	f.timeOffset = offset
	f.timeRatio = ratio
	f.Save()
}

// AddEvent appends event into frame.
func (f *Frame) AddEvent(e *Event) {
	if e.IsRoot {
		if f.Roots.AddOne(e.Hash(), e.Creator) {
			f.Save()
		}
	} else {
		if f.Events.AddOne(e.Hash(), e.Creator) {
			f.Save()
		}
	}
}

// ToWire converts to proto.Message.
func (f *Frame) ToWire() *wire.Frame {
	return &wire.Frame{
		Index:      uint32(f.Index),
		Events:     f.Events.ToWire(),
		Roots:      f.Roots.ToWire(),
		TimeOffset: uint64(f.timeOffset),
		TimeRatio:  uint64(f.timeRatio),
	}
}

// WireToFrame converts from wire.
func WireToFrame(w *wire.Frame) *Frame {
	if w == nil {
		return nil
	}
	return &Frame{
		Index:      idx.Frame(w.Index),
		Events:     WireToEventsByPeer(w.Events),
		Roots:      WireToEventsByPeer(w.Roots),
		timeOffset: inter.Timestamp(w.TimeOffset),
		timeRatio:  inter.Timestamp(w.TimeRatio),
	}
}

/*
 * Poset's methods:
 */

func (p *Poset) setFrameSaving(f *Frame) {
	f.save = func() {
		p.store.SetFrame(p.SuperFrameN, f)
	}
}
