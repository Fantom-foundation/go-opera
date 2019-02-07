package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// TODO: make State internal

// State is a poset current state.
type State struct {
	CurrentFrameN uint64
	Genesis       common.Hash
	TotalCap      uint64
}

func (p *Poset) bootstrap() {
	// restore state
	p.state = p.store.GetState()
	if p.state == nil {
		panic("Apply genesis for store first")
	}
	// TODO: restore all others from store.
}

// State saves current State
func (p *Poset) saveState() {
	p.store.SetState(p.state)
}
