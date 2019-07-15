package posposet

import (
	"sync/atomic"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// TODO: make Frame internal

// Frame is a consensus tables for frame.
type Frame struct {
	Index    idx.Frame
	Events   EventsByPeer
	Roots    EventsByPeer
	Balances hash.Hash // TODO: move to super-frame

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
	return e.LamportTime*f.timeRatio + f.timeOffset
}

// SetTimes set new timeOffset and new TimeRatio.
func (f *Frame) SetTimes(offset, ratio inter.Timestamp) {
	f.timeOffset = offset
	f.timeRatio = ratio
	f.Save()
}

// AddRoot appends root-event into frame.
func (f *Frame) AddRoot(e *Event) {
	changed := f.Events.AddOne(e.Hash(), e.Creator)
	changed = f.Roots.AddOne(e.Hash(), e.Creator) || changed
	if changed {
		f.Save()
	}
}

// AddEvent appends event into frame.
func (f *Frame) AddEvent(e *Event) {
	if f.Events.AddOne(e.Hash(), e.Creator) {
		f.Save()
	}
}

// SetBalances saves PoS-balances state.
func (f *Frame) SetBalances(balances hash.Hash) bool {
	if f.Balances != balances {
		f.Balances = balances
		f.Save()
		return true
	}
	return false
}

// ToWire converts to proto.Message.
func (f *Frame) ToWire() *wire.Frame {
	return &wire.Frame{
		Index:      uint32(f.Index),
		Balances:   f.Balances.Bytes(),
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
		Balances:   hash.FromBytes(w.Balances),
		timeOffset: inter.Timestamp(w.TimeOffset),
		timeRatio:  inter.Timestamp(w.TimeRatio),
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

// LastSuperFrame returns super-frame and list of peers
func (p *Poset) LastSuperFrame() (idx.SuperFrame, []hash.Peer) {
	n := idx.SuperFrame(atomic.LoadUint64((*uint64)(&p.SuperFrameN)))

	return n, p.SuperFrame(n)
}

// SuperFrame returns list of peers for n super-frame
func (p *Poset) SuperFrame(n idx.SuperFrame) []hash.Peer {
	members := p.store.GetMembers(n)

	addrs := make([]hash.Peer, 0, len(members))

	for member := range members {
		addrs = append(addrs, member)
	}

	return addrs
}
