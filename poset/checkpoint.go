package poset

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/poset/election"
	"github.com/Fantom-foundation/go-lachesis/vector"
)

// Checkpoint is for persistent storing.
type Checkpoint struct {
	// fields can change only after a frame is decided
	LastDecidedFrame idx.Frame
	LastBlockN       idx.Block
	LastAtropos      hash.Event
	NextValidators   pos.Validators
	StateHash        common.Hash
}

/*
 * Poset's methods:
 */

// State saves Checkpoint.
func (p *Poset) saveCheckpoint() {
	p.store.SetCheckpoint(p.Checkpoint)
}

// Bootstrap restores poset's state from store.
func (p *Poset) Bootstrap(applyBlock inter.ApplyBlockFn) {
	if p.Checkpoint != nil {
		return
	}
	// block handler must be set before p.handleElection
	p.applyBlock = applyBlock

	// restore Checkpoint
	p.Checkpoint = p.store.GetCheckpoint()
	if p.Checkpoint == nil {
		p.Log.Crit("Apply genesis for store first")
	}

	// restore current epoch
	p.loadEpoch()
	p.vecClock = vector.NewIndex(p.dag.IndexConfig, p.Validators, p.store.epochTable.VectorIndex, func(id hash.Event) *inter.EventHeaderData {
		return p.input.GetEventHeader(p.EpochN, id)
	})
	p.election = election.New(p.Validators, p.LastDecidedFrame+1, p.rootObservesRoot)

	// events reprocessing
	p.handleElection(nil)
}
