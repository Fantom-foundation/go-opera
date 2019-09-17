package poset

import (
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

const (
	firstFrame = idx.Frame(1)
	firstEpoch = idx.Epoch(1)
)

type epochState struct {
	// stored values
	// these values change only after a change of epoch
	EpochN    idx.Epoch
	PrevEpoch GenesisState
	Members   pos.Members
}

func (p *Poset) loadEpoch() {
	p.epochState = *p.store.GetEpoch()
}

// GetEpoch returns current epoch num to 3rd party.
func (p *Poset) GetEpoch() idx.Epoch {
	return idx.Epoch(atomic.LoadUint32((*uint32)(&p.EpochN)))
}

// EpochMembers returns members of current epoch.
func (p *Poset) GetMembers() pos.Members {
	return p.Members.Copy()
}

// GetEpochMembers atomically returns members of current epoch, and the epoch.
func (p *Poset) GetEpochMembers() (pos.Members, idx.Epoch) {
	return p.GetMembers(), p.GetEpoch() // TODO atomic
}

// Due to a fork, there may be many roots B with the same slot,
// but forkless caused may be only one of them (if no more than 1/3n are Byzantine), with a specific hash.
func (p *Poset) rootForklessCausesRoot(a hash.Event, bNode common.Address, bFrame idx.Frame) *hash.Event {
	var bHash *hash.Event
	p.store.ForEachRootFrom(bFrame, bNode, func(f idx.Frame, from common.Address, b hash.Event) bool {
		if f != bFrame {
			p.Log.Crit("frame mismatch")
		}
		if from != bNode {
			p.Log.Crit("node mismatch")
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
		p.Log.Crit("no genesis found")
	}
	return epoch.PrevEpoch.Hash()
}
