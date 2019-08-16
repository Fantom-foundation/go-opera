package poset

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/poset/election"
	"github.com/Fantom-foundation/go-lachesis/src/vector"
)

// checkpoint is for persistent storing.
type checkpoint struct {
	// fields can change only after a frame is decided
	LastDecidedFrame  idx.Frame
	LastBlockN        idx.Block
	LastFiWitness     hash.Event
	LastConsensusTime inter.Timestamp
	NextMembers       pos.Members
	StateHash         hash.Hash
}

/*
 * Poset's methods:
 */

// State saves checkpoint.
func (p *Poset) saveCheckpoint() {
	p.store.SetCheckpoint(p.checkpoint)
}

// Bootstrap restores checkpoint from store.
func (p *Poset) Bootstrap(applyBlock inter.ApplyBlockFn) {
	if p.checkpoint != nil {
		return
	}
	// block handler must be set before p.handleElection
	p.applyBlock = applyBlock

	// restore checkpoint
	p.checkpoint = p.store.GetCheckpoint()
	if p.checkpoint == nil {
		p.Fatal("Apply genesis for store first")
	}

	// restore current super-frame
	p.loadSuperFrame()
	p.seeVec = vector.NewIndex(p.Members, p.store.epochTable.VectorIndex)
	p.election = election.New(p.Members, p.LastDecidedFrame+1, p.rootStronglySeeRoot)

	// events reprocessing
	p.handleElection(nil)
}
