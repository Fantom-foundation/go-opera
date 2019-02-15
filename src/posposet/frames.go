package posposet

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

/*
 * Poset's methods:
 */

// lastNodeFrame returns frame of node's last root.
// Returns nil if frame is too old or or does not exist.
func (p *Poset) lastNodeFrame(node common.Address) *Frame {
	for i := len(p.frames); i > 0; i-- {
		n := p.state.LastFinishedFrameN + uint64(i)
		f := p.frame(n, false)
		if f != nil && f.HasNodeEvent(node) {
			return f
		}
	}
	return nil
}

// frame finds or creates frame.
func (p *Poset) frame(n uint64, orCreate bool) *Frame {
	if n < p.state.LastFinishedFrameN {
		panic(fmt.Errorf("Too old frame%d is requested", n))
	}
	// return ephemeral
	if n == 0 {
		return &Frame{
			Index:    0,
			Balances: p.state.Genesis,
		}
	}
	// return existing
	f := p.frames[n]
	if f == nil {
		if !orCreate {
			return nil
		}
		// create new frame
		f = &Frame{
			Index:     n,
			FlagTable: FlagTable{},
			NonRoots:  Roots{},
		}
	}

	f.save = p.saveFuncForFrame(f)
	p.frames[n] = f

	return f
}
