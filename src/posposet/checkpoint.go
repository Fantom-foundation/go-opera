package posposet

import (
	"sync/atomic"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/election"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// checkpoint is for persistent storing.
type checkpoint struct {
	SuperFrameN        uint64
	lastFinishedFrameN uint64
	LastBlockN         uint64
	Genesis            hash.Hash
	TotalCap           uint64
}

// LastFinishedFrameN is a getter of lastFinishedFrameN.
func (cp *checkpoint) LastFinishedFrameN() uint64 {
	return atomic.LoadUint64(&cp.lastFinishedFrameN)
}

// LastFinishedFrame is a setter of lastFinishedFrameN.
func (cp *checkpoint) LastFinishedFrame(N uint64) {
	atomic.StoreUint64(&cp.lastFinishedFrameN, N)
}

// ToWire converts to proto.Message.
func (cp *checkpoint) ToWire() *wire.Checkpoint {
	return &wire.Checkpoint{
		SuperFrameN:        cp.SuperFrameN,
		LastFinishedFrameN: cp.LastFinishedFrameN(),
		LastBlockN:         cp.LastBlockN,
		Genesis:            cp.Genesis.Bytes(),
		TotalCap:           cp.TotalCap,
	}
}

// WireToState converts from wire.
func WireToCheckpoint(w *wire.Checkpoint) *checkpoint {
	if w == nil {
		return nil
	}
	return &checkpoint{
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
	p.members = p.store.GetMembers(p.SuperFrameN)
	p.election = election.NewElection(p.members, election.Amount(p.members.TotalStake()), election.Amount(p.getSuperMajority()), election.IdxFrame(p.checkpoint.lastFinishedFrameN+1), p.rootStronglySeeRoot)

	// restore frames
	for n := p.LastFinishedFrameN(); true; n++ {
		if f := p.store.GetFrame(n, p.SuperFrameN); f != nil {
			p.frames[n] = f
		} else if n > 0 {
			break
		}
	}
	// recalc in case there was a interrupted consensus
	start := p.frame(p.LastFinishedFrameN(), true)
	p.reconsensusFromFrame(p.LastFinishedFrameN()+1, start.Balances)
}

// GetGenesisHash is a genesis getter.
func (p *Poset) GetGenesisHash() hash.Hash {
	return p.Genesis
}

// GenesisHash calcs hash of genesis balances.
func genesisHash(balances map[hash.Peer]uint64) hash.Hash {
	s := NewMemStore()
	defer s.Close()

	if err := s.ApplyGenesis(balances); err != nil {
		logger.Get().Fatal(err)
	}

	return s.GetCheckpoint().Genesis
}
