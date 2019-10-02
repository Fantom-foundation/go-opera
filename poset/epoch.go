package poset

import (
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
)

const (
	firstFrame = idx.Frame(1)
	firstEpoch = idx.Epoch(1)
)

type epochState struct {
	// stored values
	// these values change only after a change of epoch
	EpochN     idx.Epoch
	PrevEpoch  GenesisState
	Validators pos.Validators
}

func (p *Poset) loadEpoch() {
	p.epochState = *p.store.GetEpoch()
	p.store.RecreateEpochDb(p.epochState.EpochN)
}

// GetEpoch returns current epoch num to 3rd party.
func (p *Poset) GetEpoch() idx.Epoch {
	return idx.Epoch(atomic.LoadUint32((*uint32)(&p.EpochN)))
}

// EpochValidators returns validators of current epoch.
func (p *Poset) GetValidators() pos.Validators {
	return p.Validators.Copy()
}

// GetEpochValidators atomically returns validators of current epoch, and the epoch.
func (p *Poset) GetEpochValidators() (pos.Validators, idx.Epoch) {
	return p.GetValidators(), p.GetEpoch() // TODO atomic
}

// rootObservesRoot returns hash of root B, if root B forkless causes root A.
// Due to a fork, there may be many roots B with the same slot,
// but A may be forkless caused only by one of them (if no more than 1/3n are Byzantine), with a specific hash.
func (p *Poset) rootObservesRoot(a hash.Event, bCreator common.Address, bFrame idx.Frame) *hash.Event {
	var bHash *hash.Event
	p.store.ForEachRootFrom(bFrame, bCreator, func(f idx.Frame, from common.Address, b hash.Event) bool {
		if f != bFrame || from != bCreator {
			p.Log.Crit("Inconsistent DB iteration")
		}
		if p.vecClock.ForklessCause(a, b) {
			bHash = &b
			return false
		}
		return true
	})

	return bHash
}

// GetGenesisHash is a genesis getter.
func (p *Poset) GetGenesisHash() common.Hash {
	epoch := p.store.GetGenesis()
	if epoch == nil {
		p.Log.Crit("No genesis found")
	}
	return epoch.PrevEpoch.Hash()
}
