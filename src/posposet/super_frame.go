package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/ordering"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/election"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/seeing"
)

type superFrame struct {
	// state
	frames  map[idx.Frame]*Frame
	members internal.Members

	// election votes
	election *election.Election

	strongly *seeing.Strongly
}

func (p *Poset) initSuperFrame() {
	p.members = p.store.GetMembers(p.SuperFrameN)

	p.strongly = seeing.New(p.members)
	p.frames = make(map[idx.Frame]*Frame)
	for n := idx.Frame(1); true; n++ {
		frame := p.store.GetFrame(n, p.SuperFrameN)
		if frame == nil {
			break
		}
		p.frames[n] = frame

		cached := make(map[hash.Event]*inter.Event)
		orderThenCache := ordering.EventBuffer(ordering.Callback{
			Process: func(e *inter.Event) {
				p.strongly.Cache(e)
				cached[e.Hash()] = e
			},
			Drop: func(e *inter.Event, err error) {
				p.Fatal(err)
			},
			Exists: func(e hash.Event) *inter.Event {
				return cached[e]
			},
		})

		for _, ee := range frame.Events {
			for e := range ee {
				event := p.GetEvent(e)
				orderThenCache(event.Event)
			}
		}

	}

	p.election = election.New(p.members, p.LastDecidedFrameN+1, p.rootStronglySeeRoot)
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
