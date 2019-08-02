package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// TODO: make Frame internal

// Frame is a consensus tables for frame.
type Frame struct {
	Index idx.Frame
	Roots EventsByPeer

	TimeOffset inter.Timestamp
	TimeRatio  inter.Timestamp

	save func() `rlp:"-"`
}

// Save calls .save() if set.
func (f *Frame) Save() {
	if f.save != nil {
		f.save()
	}
}

// GetConsensusTimestamp calc consensus timestamp for given event.
func (f *Frame) GetConsensusTimestamp(e *Event) inter.Timestamp {
	return inter.Timestamp(e.Lamport)*f.TimeOffset + f.TimeRatio
}

// SetTimes set new timeOffset and new TimeRatio.
func (f *Frame) SetTimes(offset, ratio inter.Timestamp) {
	f.TimeOffset = offset
	f.TimeRatio = ratio
	f.Save()
}

// AddEvent appends event into frame.
func (f *Frame) AddEvent(e *Event) {
	if e.IsRoot {
		if f.Roots.AddOne(e.Hash(), e.Creator) {
			f.Save()
		}
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
