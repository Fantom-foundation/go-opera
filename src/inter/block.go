package inter

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// Block is a chain block.
type Block struct {
	Index  idx.Block
	Events hash.OrderedEvents
}

// NewBlock makes main chain block from topological ordered events.
func NewBlock(index idx.Block, ordered Events) *Block {
	events := make(hash.OrderedEvents, len(ordered))
	for i, e := range ordered {
		events[i] = e.Hash()
	}

	return &Block{
		Index:  index,
		Events: events,
	}
}
