package posposet

import (
	"sync/atomic"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// TODO: make State internal

// State is a current poset state.
type State struct {
	SuperFrameN        uint64
	lastFinishedFrameN uint64
	LastBlockN         uint64
	Genesis            hash.Hash
	TotalCap           uint64
}

func (s *State) LastFinishedFrameN() uint64 {
	return atomic.LoadUint64(&s.lastFinishedFrameN)
}

func (s *State) LastFinishedFrame(N uint64) {
	atomic.StoreUint64(&s.lastFinishedFrameN, N)
}

// ToWire converts to proto.Message.
func (s *State) ToWire() *wire.State {
	return &wire.State{
		SuperFrameN:        s.SuperFrameN,
		LastFinishedFrameN: s.LastFinishedFrameN(),
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
		SuperFrameN:        w.SuperFrameN,
		lastFinishedFrameN: w.LastFinishedFrameN,
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
		p.Fatal("Apply genesis for store first")
	}
	// restore frames
	for n := p.state.LastFinishedFrameN(); true; n++ {
		if f := p.store.GetFrame(n); f != nil {
			p.frames[n] = f
		} else if n > 0 {
			break
		}
	}
	// recalc in case there was a interrupted consensus
	start := p.frame(p.state.LastFinishedFrameN(), true)
	p.reconsensusFromFrame(p.state.LastFinishedFrameN()+1, start.Balances)
}

// GetGenesisHash is a genesis getter.
func (p *Poset) GetGenesisHash() hash.Hash {
	return p.state.Genesis
}

// GenesisHash calcs hash of genesis balances.
func genesisHash(balances map[hash.Peer]uint64) hash.Hash {
	s := NewMemStore()
	defer s.Close()

	if err := s.ApplyGenesis(balances); err != nil {
		logger.Get().Fatal(err)
	}

	return s.GetState().Genesis
}
