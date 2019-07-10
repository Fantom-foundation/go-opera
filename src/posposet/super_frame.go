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
	p.election = election.New(p.members, p.checkpoint.lastFinishedFrameN+1, p.stronglySee)

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

func (p *Poset) stronglySee(a hash.Event, b hash.Peer, f idx.Frame) *hash.Event {
	frame := p.frame(f, false)
	if frame == nil {
		return nil
	}

	for r := range frame.Events[b] {
		if p.strongly.See(a, r) {
			return &r
		}
	}

	return nil
}
