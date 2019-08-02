package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/posposet/vectorindex"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
	"sync/atomic"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/election"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
)

const (
	// SuperFrameLen is a count of FW per super-frame.
	SuperFrameLen uint = 100

	firstFrame = idx.Frame(1)
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
	Genesis        GenesisState
	SfWitnessCount uint
	Balances       hash.Hash
	Members        internal.Members
	NextMembers    internal.Members

	frames map[idx.Frame]*Frame `rlp:"-"`

	// election votes
	election *election.Election `rlp:"-"`

	vi *vectorindex.Vindex `rlp:"-"`
}

func (p *Poset) loadSuperFrame() {
	p.superFrame = *p.store.GetSuperFrame(p.SuperFrameN)
	p.NextMembers = p.Members.Top()
	p.vi = vectorindex.New(p.Members, p.store.table.VectorIndex)
	p.election = election.New(p.Members, firstFrame, p.rootStronglySeeRoot)
	p.frames = make(map[idx.Frame]*Frame)

	// events reprocessing
	// TODO store frames in DB, so we won't need to re-process the whole epoch
	/*
		toReprocess := hash.EventsSet{}
		orderThenReprocess := ordering.EventBuffer(ordering.Callback{
			Process: func(e *inter.Event) {
				p.consensus(e)
				delete(toReprocess, e.Hash())
			},

			Drop: func(e *inter.Event, err error) {
				p.Fatal(err.Error() + ", so rejected")
			},

			Exists: func(h hash.Event) *inter.Event {
				if toReprocess.Contains(h) {
					return nil
				}
				return p.input.GetEvent(h)
			},
		})

		for _, f := range []func(e hash.Event){
			func(e hash.Event) { // save event list first
				toReprocess.Add(e)
			},
			func(e hash.Event) { // then process events
				orderThenReprocess(p.input.GetEvent(e))
			},
		} {
			// all events from frames of SF
			for n := firstFrame; true; n++ {
				frame := p.store.GetFrame(p.SuperFrameN, n)
				if frame == nil {
					break
				}
				for _, src := range []EventsByPeer{frame.Events, frame.Roots} {
					for _, ee := range src {
						for e := range ee {
							f(e)
						}
					}
				}
			}
		}
	*/
}

func (p *Poset) nextEpoch(fiWitness hash.Event) {
	// new genesis state
	p.Genesis.Time = p.LastConsensusTime
	p.Genesis.Epoch = p.SuperFrameN
	p.Genesis.LastFiWitness = fiWitness
	p.Genesis.StateHash = p.superFrame.Balances

	// new members list
	p.Members = p.NextMembers
	p.NextMembers = p.Members.Top()

	// reset internal epoch state
	p.frames = make(map[idx.Frame]*Frame)

	p.vi.Reset(p.store.table.VectorIndex) // TODO is this DB pruned after new epoch?
	p.election.Reset(p.Members, firstFrame)

	// set new epoch
	p.SuperFrameN++
	p.store.SetSuperFrame(p.SuperFrameN, &p.superFrame)
}

// CurrentSuperFrameN returns current super-frame num to 3rd party.
func (p *Poset) CurrentSuperFrameN() idx.SuperFrame {
	return idx.SuperFrame(atomic.LoadUint64((*uint64)(&p.SuperFrameN)))
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
	frame, ok := p.frames[bFrame]
	if !ok { // not known frame for B
		return nil
	}

	for b := range frame.Roots[bNode] {
		if p.vi.StronglySee(a, b) {
			return &b
		}
	}

	return nil
}
