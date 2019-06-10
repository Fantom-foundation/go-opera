package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// GetEventBlock returns block includes event.
func (p *Poset) GetEventBlock(e hash.Event) *inter.Block {
	num := p.store.GetEventBlockNum(e)
	if num == nil {
		return nil
	}

	return p.store.GetBlock(*num)
}
