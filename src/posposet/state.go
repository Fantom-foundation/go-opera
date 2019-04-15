package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// TODO: make State internal

// State is a current poset state.
type State struct {
	LastFinishedFrameN uint64
	LastBlockN         uint64
	Genesis            hash.Hash
	TotalCap           uint64
}

// ToWire converts to proto.Message.
func (s *State) ToWire() *wire.State {
	return &wire.State{
		LastFinishedFrameN: s.LastFinishedFrameN,
		LastBlockN:         s.LastBlockN,
		Genesis:            s.Genesis.Bytes(),
		TotalCap:           s.TotalCap,
	}
}

// WireToState converts from wire.
func WireToState(w *wire.State) *State {
	if w == nil {
		return nil
	}
	return &State{
		LastFinishedFrameN: w.LastFinishedFrameN,
		LastBlockN:         w.LastBlockN,
		Genesis:            hash.FromBytes(w.Genesis),
		TotalCap:           w.TotalCap,
	}
}

/*
 * Poset's methods:
 */

// State saves current state.
func (p *Poset) saveState() {
	p.store.SetState(p.state)
}

// Bootstrap restores current state from store.
func (p *Poset) Bootstrap() {
	if p.state != nil {
		return
	}
	// restore state
	p.state = p.store.GetState()
	if p.state == nil {
		panic("Apply genesis for store first")
	}
	// restore frames
	for n := p.state.LastFinishedFrameN; true; n++ {
		if f := p.store.GetFrame(n); f != nil {
			p.frames[n] = f
		} else if n > 0 {
			break
		}
	}
	// recalc in case there was a interrupted consensus
	p.reconsensusFromFrame(p.state.LastFinishedFrameN + 1)
}

func GenesisHash(balances map[hash.Peer]uint64) hash.Hash {
	s := NewMemStore()
	defer s.Close()

	if err := s.ApplyGenesis(balances); err != nil {
		panic(err)
	}

	return s.GetState().Genesis
}
