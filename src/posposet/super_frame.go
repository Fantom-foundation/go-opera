package posposet

import (
	"sync/atomic"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/election"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/seeing"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

const (
	SuperFrameLen int = 100

	firstFrame = idx.Frame(1)
)

type superFrame struct {
	sfWitnessCount int
	frames         map[idx.Frame]*Frame
	balances       hash.Hash
	members        internal.Members
	nextMembers    internal.Members

	// election votes
	election *election.Election

	strongly *seeing.Strongly
}

func (sf *superFrame) ToWire() *wire.SuperFrame {
	return &wire.SuperFrame{
		Balances: sf.balances.Bytes(),
		Members:  sf.members.ToWire(),
	}
}

func WireToSuperFrame(w *wire.SuperFrame) (sf *superFrame) {
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

	p.strongly = seeing.New(p.members.NewCounter)
	p.election = election.New(p.members, firstFrame, p.rootStronglySeeRoot)
	p.nextMembers = internal.Members{}
	p.frames = make(map[idx.Frame]*Frame)

	// events reprocessing
	toReload := hash.Events{}
	for n := firstFrame; true; n++ {
		frame := p.store.GetFrame(p.SuperFrameN, n)
		if frame == nil {
			break
		}
		for _, src := range []EventsByPeer{frame.Events, frame.Roots} {
			for _, ee := range src {
				toReload.Add(ee.Slice()...)
			}
		}
	}

	for e := range toReload {
		p.PushEvent(e)
	}
}

func (p *Poset) nextSuperFrame() {
	p.members = p.nextMembers
	p.nextMembers = internal.Members{}

	p.frames = make(map[idx.Frame]*Frame)

	p.strongly.Reset()
	p.election.Reset(p.members, firstFrame)

	p.SuperFrameN += 1
	p.store.SetSuperFrame(p.SuperFrameN, &p.superFrame)
}

// SuperFrame returns list of peers for n super-frame.
// If req==0 returns last.
func (p *Poset) SuperFramePeers(req idx.SuperFrame) (n idx.SuperFrame, members []hash.Peer) {
	if req == idx.SuperFrame(0) {
		n = idx.SuperFrame(atomic.LoadUint64((*uint64)(&p.SuperFrameN)))
	} else {
		n = req
	}

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
		if p.strongly.See(a, b) {
			return &b
		}
	}

	return nil
}
