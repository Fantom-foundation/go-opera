package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/posposet/vectorindex"
	"sync/atomic"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/ordering"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/election"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

const (
	// SuperFrameLen is a count of FW per super-frame.
	SuperFrameLen int = 100

	firstFrame = idx.Frame(1)
)

// state of previous Epoch
type genesisState struct {
	epoch         idx.SuperFrame
	time          inter.Timestamp // consensus time of the last fiWitness
	lastFiWitness hash.Event
	stateHash     hash.Hash // hash of txs state. TBD
}

func (g *genesisState) Hash() hash.Hash {
	// TODO
	/*hasher := sha3.New256()
	rlp.Encode(hasher, g)
	return hash.Hash(hasher.Sum(nil))*/
	return hash.Hash{}
}

type superFrame struct {
	genesis genesisState

	sfWitnessCount int
	frames         map[idx.Frame]*Frame
	balances       hash.Hash
	members        internal.Members
	nextMembers    internal.Members

	// election votes
	election *election.Election

	vi *vectorindex.Vindex
}

// ToWire converts to protobuf message.
func (sf *superFrame) ToWire() *wire.SuperFrame {
	return &wire.SuperFrame{
		Balances: sf.balances.Bytes(),
		Members:  sf.members.ToWire(),
	}
}

// wireToSuperFrame converts from protobuf message.
func wireToSuperFrame(w *wire.SuperFrame) (sf *superFrame) {
	if w == nil {
		return
	}

	sf = &superFrame{
		balances: hash.FromBytes(w.Balances),
		members:  internal.WireToMembers(w.Members),
	}

	return
}

func (p *Poset) loadSuperFrame() {
	p.superFrame = *p.store.GetSuperFrame(p.SuperFrameN)
	p.nextMembers = p.members.Top()
	p.vi = vectorindex.New(p.members)
	p.election = election.New(p.members, firstFrame, p.rootStronglySeeRoot)
	p.frames = make(map[idx.Frame]*Frame)

	// events reprocessing
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
}

func (p *Poset) nextEpoch(fiWitness hash.Event) {
	// new genesis state
	p.genesis.time = p.LastConsensusTime
	p.genesis.epoch = p.SuperFrameN
	p.genesis.lastFiWitness = fiWitness
	p.genesis.stateHash = p.superFrame.balances

	// new members list
	p.members = p.nextMembers
	p.nextMembers = p.members.Top()

	// reset internal epoch state
	p.frames = make(map[idx.Frame]*Frame)

	p.vi.Reset()
	p.election.Reset(p.members, firstFrame)

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

	for m := range sf.members {
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
