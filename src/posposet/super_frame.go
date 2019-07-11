package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
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
	p.election = election.New(p.members, p.checkpoint.lastFinishedFrameN+1, p.rootStronglySeeRoot)

	// restore frames
	p.frames = make(map[idx.Frame]*Frame)
	for n := p.LastFinishedFrameN(); true; n++ {
		if f := p.store.GetFrame(n, p.SuperFrameN); f != nil {
			p.frames[n] = f
		} else if n > 0 {
			break
		}
	}
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
