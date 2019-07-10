package posposet

import (
	"sync/atomic"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// checkpoint is for persistent storing.
type checkpoint struct {
	SuperFrameN        idx.SuperFrame
	lastFinishedFrameN idx.Frame
	LastBlockN         idx.Block
	Genesis            hash.Hash
	TotalCap           inter.Stake
}

// LastFinishedFrameN is a getter of lastFinishedFrameN.
func (cp *checkpoint) LastFinishedFrameN() idx.Frame {
	return idx.Frame(atomic.LoadUint32((*uint32)(&cp.lastFinishedFrameN)))
}

// LastFinishedFrame is a setter of lastFinishedFrameN.
func (cp *checkpoint) LastFinishedFrame(N idx.Frame) {
	atomic.StoreUint32((*uint32)(&cp.lastFinishedFrameN), uint32(N))
}

// ToWire converts to proto.Message.
func (cp *checkpoint) ToWire() *wire.Checkpoint {
	return &wire.Checkpoint{
		SuperFrameN:        uint64(cp.SuperFrameN),
		LastFinishedFrameN: uint32(cp.LastFinishedFrameN()),
		LastBlockN:         uint64(cp.LastBlockN),
		Genesis:            cp.Genesis.Bytes(),
		TotalCap:           uint64(cp.TotalCap),
	}
}

// WireToState converts from wire.
func WireToCheckpoint(w *wire.Checkpoint) *checkpoint {
	if w == nil {
		return nil
	}
	return &checkpoint{
		SuperFrameN:        idx.SuperFrame(w.SuperFrameN),
		lastFinishedFrameN: idx.Frame(w.LastFinishedFrameN),
		LastBlockN:         idx.Block(w.LastBlockN),
		Genesis:            hash.FromBytes(w.Genesis),
		TotalCap:           inter.Stake(w.TotalCap),
	}
}

/*
 * Poset's methods:
 */

// State saves checkpoint.
func (p *Poset) saveCheckpoint() {
	p.store.SetCheckpoint(p.checkpoint)
}

// Bootstrap restores checkpoint from store.
func (p *Poset) Bootstrap() {
	if p.checkpoint != nil {
		return
	}
	// restore checkpoint
	p.checkpoint = p.store.GetCheckpoint()
	if p.checkpoint == nil {
		p.Fatal("Apply genesis for store first")
	}
	// restore current super-frame
	p.initSuperFrame()

	// recalc in case there was a interrupted consensus
	start := p.frame(p.LastFinishedFrameN(), true)
	p.reloadFromFrame(p.LastFinishedFrameN()+1, start.Balances)
}

// reloadFromFrame reconsensus events from non-finished frames.
func (p *Poset) reloadFromFrame(start idx.Frame, newBalance hash.Hash) {
	// TODO: implement it
	/*	stop := p.frameNumLast()
		var all inter.Events
		// foreach stale frame
		for n := start; n <= stop; n++ {
			frame := p.frames[n]
			// extract events
			for e := range frame.Events.Each() {
				all = append(all, p.input.GetEvent(e))
			}
			// and replace stale frame with blank
			p.frames[n] = &Frame{
				Index:    n,
				Events:   EventsByPeer{},
				Roots:    EventsByPeer{},
				Balances: newBalance,
			}
		}
		// recalc consensus (without frame saving)
		for _, e := range all.ByParents() {
			p.consensus(e)
		}

		// save fresh frame
		for n := start; n <= stop; n++ {
			frame := p.frames[n]

			p.setFrameSaving(frame)
			frame.Save()
		}
	*/
}

// GetGenesisHash is a genesis getter.
func (p *Poset) GetGenesisHash() hash.Hash {
	return p.Genesis
}

// GenesisHash calcs hash of genesis balances.
func genesisHash(balances map[hash.Peer]inter.Stake) hash.Hash {
	s := NewMemStore()
	defer s.Close()

	if err := s.ApplyGenesis(balances); err != nil {
		logger.Get().Fatal(err)
	}

	return s.GetCheckpoint().Genesis
}
