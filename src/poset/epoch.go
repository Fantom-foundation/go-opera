package poset

import (
	"fmt"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

const (
	firstFrame = idx.Frame(1)
	firstEpoch = idx.Epoch(1)
)

// state of previous Epoch
type GenesisState struct {
	Epoch       idx.Epoch
	Time        inter.Timestamp // consensus time of the last atropos
	LastAtropos hash.Event
	StateHash   common.Hash // hash of txs state
}

func (g *GenesisState) Hash() common.Hash {
	hasher := sha3.New256()
	if err := rlp.Encode(hasher, g); err != nil {
		panic(err)
	}
	return hash.FromBytes(hasher.Sum(nil))
}

func (g *GenesisState) EpochName() string {
	return fmt.Sprintf("epoch%d", g.Epoch)
}

type epoch struct {
	// stored values
	// these values change only after a change of epoch
	EpochN    idx.Epoch
	PrevEpoch GenesisState
	Members   pos.Members
}

func (p *Poset) loadEpoch() {
	p.epoch = *p.store.GetEpoch()
}

func (p *Poset) nextEpoch(atropos hash.Event) {
	// new PrevEpoch state
	p.PrevEpoch.Time = p.LastConsensusTime
	p.PrevEpoch.Epoch = p.EpochN
	p.PrevEpoch.LastAtropos = atropos
	p.PrevEpoch.StateHash = p.checkpoint.StateHash

	// new members list
	p.Members = p.NextMembers.Top()
	p.NextMembers = p.Members.Copy()

	// reset internal epoch DB
	p.store.recreateEpochDb()

	// reset election & vectorindex
	p.vecClock.Reset(p.Members, p.store.epochTable.VectorIndex, func(id hash.Event) *inter.EventHeaderData {
		return p.input.GetEventHeader(p.EpochN, id)
	}) // this DB is pruned after .pruneTempDb()
	p.election.Reset(p.Members, firstFrame)
	p.LastDecidedFrame = 0

	// move to new epoch
	p.EpochN++

	// commit
	p.store.SetEpoch(&p.epoch)
	p.saveCheckpoint()
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

// rootForklessCausesRoot returns hash of root B, if root A forkless causes root B.
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
