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

// NewBlock makes block from topological ordered events.
func NewBlock(index idx.Block, time Timestamp, ordered Events) *Block {
	events := make(hash.Events, len(ordered))
	for i, e := range ordered {
		events[i] = e.Hash()
	}

	return &Block{
		Index:  index,
		Time:   time,
		Events: events,
	}
}
