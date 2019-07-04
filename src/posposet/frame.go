package posposet

import (
	"sync/atomic"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// TODO: make Frame internal

// Frame is a consensus tables for frame.
type Frame struct {
	Index            uint64
	FlagTable        FlagTable
	ClothoCandidates EventsByPeer
	Atroposes        TimestampsByEvent
	Balances         hash.Hash

	save func()
}

// Save calls .save() if set.
func (f *Frame) Save() {
	if f.save != nil {
		f.save()
	}
}

// AddRootsOf appends known roots for event.
func (f *Frame) AddRootsOf(event hash.Event, roots EventsByPeer) {
	if f.FlagTable[event] == nil {
		f.FlagTable[event] = EventsByPeer{}
	}
	if f.FlagTable[event].Add(roots) {
		f.Save()
	}
}

// AddClothoCandidate adds event into ClothoCandidates list.
func (f *Frame) AddClothoCandidate(event hash.Event, creator hash.Peer) {
	if f.ClothoCandidates.AddOne(event, creator) {
		f.Save()
	}
}

// SetAtropos makes Atropos from Clotho and consensus time.
func (f *Frame) SetAtropos(clotho hash.Event, consensusTime inter.Timestamp) {
	if t, ok := f.Atroposes[clotho]; ok && t == consensusTime {
		return
	}
	f.Atroposes[clotho] = consensusTime
	f.Save()
}

// GetRootsOf returns known roots of event. For read only, please.
func (f *Frame) GetRootsOf(event hash.Event) EventsByPeer {
	return f.FlagTable[event]
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
		Index:            f.Index,
		FlagTable:        f.FlagTable.ToWire(),
		ClothoCandidates: f.ClothoCandidates.ToWire(),
		Atroposes:        f.Atroposes.ToWire(),
		Balances:         f.Balances.Bytes(),
	}
}

// WireToFrame converts from wire.
func WireToFrame(w *wire.Frame) *Frame {
	if w == nil {
		return nil
	}
	return &Frame{
		Index:            w.Index,
		FlagTable:        WireToFlagTable(w.FlagTable),
		ClothoCandidates: WireToEventsByPeer(w.ClothoCandidates),
		Atroposes:        WireToTimestampsByEvent(w.Atroposes),
		Balances:         hash.FromBytes(w.Balances),
	}
}

/*
 * Poset's methods:
 */

func (p *Poset) setFrameSaving(f *Frame) {
	f.save = func() {
		if f.Index > p.state.LastFinishedFrameN() {
			p.store.SetFrame(f, p.state.SuperFrameN)
		} else {
			p.Fatalf("frame %d is finished and should not be changed", f.Index)
		}
	}
}

// LastSuperFrame returns super-frame and list of peers
func (p *Poset) LastSuperFrame() (uint64, []hash.Peer) {
	n := atomic.LoadUint64(&p.state.SuperFrameN)

	return n, p.SuperFrame(n)
}

// SuperFrame returns list of peers for n super-frame
func (p *Poset) SuperFrame(n uint64) []hash.Peer {
	members := p.store.GetMembers(n).ToWire().Members

	addrs := make([]hash.Peer, 0, len(members))

	for _, member := range members {
		addrs = append(addrs, hash.BytesToPeer(member.Addr))
	}

	return addrs
}
