package posposet

import (
	"sync/atomic"

	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
)

const (
	// SuperFrameLen is a count of FW per super-frame.
	SuperFrameLen idx.Frame = 100

	firstFrame = idx.Frame(1)
	firstEpoch = idx.SuperFrame(1)
)

// state of previous Epoch
type GenesisState struct {
	Epoch         idx.SuperFrame
	Time          inter.Timestamp // consensus time of the last fiWitness
	LastFiWitness hash.Event
	StateHash     hash.Hash // hash of txs state. TBD
}

func (g *GenesisState) Hash() hash.Hash {
	hasher := sha3.New256()
	if err := rlp.Encode(hasher, g); err != nil {
		panic(err)
	}
	return hash.FromBytes(hasher.Sum(nil))
}

type superFrame struct {
	// stored values
	// these values change only after a change of epoch
	PrevEpoch GenesisState
	Members   internal.Members
}

func (p *Poset) loadSuperFrame() {
	p.superFrame = *p.store.GetSuperFrame(p.SuperFrameN)
}

func (p *Poset) nextEpoch(fiWitness hash.Event) {
	// new PrevEpoch state
	p.PrevEpoch.Time = p.LastConsensusTime
	p.PrevEpoch.Epoch = p.SuperFrameN
	p.PrevEpoch.LastFiWitness = fiWitness
	p.PrevEpoch.StateHash = p.checkpoint.Balances

	// new members list
	p.Members = p.NextMembers
	p.NextMembers = p.Members.Top()

	// reset internal epoch DB
	p.store.pruneTempDb()

	// reset election & vectorindex
	p.events.Reset(p.Members, p.store.epochTable.VectorIndex) // this DB is pruned after .pruneTempDb()
	p.election.Reset(p.Members, firstFrame)
	p.LastDecidedFrame = 0

	// move to new epoch
	p.SuperFrameN++

	// commit
	p.store.SetSuperFrame(p.SuperFrameN, &p.superFrame)
	p.saveCheckpoint()
}

// CurrentSuperFrameN returns current super-frame num to 3rd party.
func (p *Poset) CurrentSuperFrameN() idx.SuperFrame {
	return idx.SuperFrame(atomic.LoadUint32((*uint32)(&p.SuperFrameN)))
}

// SuperFrameMembers returns members of n super-frame.
func (p *Poset) SuperFrameMembers(n idx.SuperFrame) (members []hash.Peer) {
	sf := p.store.GetSuperFrame(n)
	if sf == nil {
		p.Fatalf("super-frame %d not found", n)
	}

	for m := range sf.Members {
		members = append(members, m)
	}

	return
}

// rootStronglySeeRoot returns hash of root B, if root A strongly sees root B.
// Due to a fork, there may be many roots B with the same slot,
// but strongly seen may be only one of them (if no more than 1/3n are Byzantine), with a specific hash.
func (p *Poset) rootStronglySeeRoot(a hash.Event, bNode hash.Peer, bFrame idx.Frame) *hash.Event {
	var bRoots hash.Events
	p.store.ForEachRootFrom(bFrame, bNode, func(f idx.Frame, from hash.Peer, b hash.Event) bool {
		if f != bFrame {
			p.Fatal()
		}
		if from != bNode {
			p.Fatal()
		}
		bRoots.Add(hash.BytesToEvent(b.Bytes()))
		return true
	})
	for _, b := range bRoots {
		if p.events.StronglySee(a, b) {
			return &b
		}
	}

	return nil
}

// GetGenesisHash is a genesis getter.
func (p *Poset) GetGenesisHash() hash.Hash {
	epoch := p.store.GetSuperFrame(firstEpoch)
	if epoch == nil {
		p.Fatal("no genesis found")
	}
	return epoch.PrevEpoch.Hash()
}
