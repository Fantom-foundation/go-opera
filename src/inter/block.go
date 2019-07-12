package inter

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

// Block is a chain block.
type Block struct {
	Index  idx.Block
	Events hash.OrderedEvents
}

// ToWire converts to proto.Message.
func (b *Block) ToWire() *wire.Block {
	if b == nil {
		return nil
	}
	return &wire.Block{
		Index:  uint64(b.Index),
		Events: b.Events.ToWire(),
	}
}

// WireToBlock converts from wire.
func WireToBlock(w *wire.Block) *Block {
	if w == nil {
		return nil
	}
	return &Block{
		Index:  idx.Block(w.Index),
		Events: hash.WireToOrderedEvents(w.Events),
	}
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
