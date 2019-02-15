package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// TODO: make State internal

// State is a current poset state.
type State struct {
	LastFinishedFrameN uint64
	Genesis            common.Hash
	TotalCap           uint64
}

/*
 * Poset's methods:
 */

// State saves current state.
func (p *Poset) saveState() {
	p.store.SetState(p.state)
}

// bootstrap restores current state from store.
func (p *Poset) bootstrap() {
	// restore state
	p.state = p.store.GetState()
	if p.state == nil {
		panic("Apply genesis for store first")
	}
}
