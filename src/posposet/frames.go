package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// FrameOfEvent returns unfinished frame where event is in.
func (p *Poset) FrameOfEvent(event hash.Event) *Frame {
	e := p.GetEventHeader(event)
	if e != nil && e.Epoch == p.SuperFrameN {
		return p.frames[e.Frame]
	}

	return nil
}

// frame finds or creates frame.
func (p *Poset) frame(n idx.Frame, orCreate bool) *Frame {
	// return ephemeral
	if n == 0 {
		return &Frame{
			Index: 0,
		}
	}

	// return existing
	f, ok := p.frames[n]
	if !ok {
		if !orCreate {
			return nil
		}
		// create new frame
		f = &Frame{
			Index: n,
			Roots: EventsByPeer{},
		}
		p.setFrameSaving(f)
		p.frames[n] = f
		f.save()
	}

	return f
}

// frameNumLast returns last frame number.
func (p *Poset) frameNumLast() idx.Frame {
	var max idx.Frame
	for n := range p.frames {
		if max < n {
			max = n
		}

	}
	return max
}
