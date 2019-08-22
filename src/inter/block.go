package inter

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

type ApplyBlockFn func(block *Block, stateHash hash.Hash, members pos.Members) (newStateHash hash.Hash, mewMembers pos.Members)

// Block is a "chain" block.
type Block struct {
	Index  idx.Block
	Time   Timestamp
	Events hash.Events
}

func (b *Block) Head() hash.Event {
	if len(b.Events) == 0 {
		return hash.ZeroEvent
	}
	return b.Events[len(b.Events)-1] // fiWitness is always a last event
}

// NewBlock makes block from topological ordered events.
func NewBlock(index idx.Block, time Timestamp, events hash.Events) *Block {
	return &Block{
		Index:  index,
		Time:   time,
		Events: events,
	}
}
