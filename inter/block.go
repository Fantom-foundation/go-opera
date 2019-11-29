package inter

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
)

// ConsensusCallbacks contains callbacks called during block processing by consensus engine
type ConsensusCallbacks struct {
	// ApplyBlock is callback type to apply the new block to the state
	ApplyBlock func(block *Block, decidedFrame idx.Frame, cheaters Cheaters) (newAppHash common.Hash, sealEpoch bool)
	// SelectValidatorsGroup is a callback type to select new validators group.
	SelectValidatorsGroup func(oldEpoch, newEpoch idx.Epoch) (newValidators *pos.Validators)
	// OnEventConfirmed is callback type to notify about event confirmation.
	OnEventConfirmed func(event *EventHeaderData, seqDepth idx.Event)
	// IsEventAllowedIntoBlock is callback type to check is event may be within block or not
	IsEventAllowedIntoBlock func(event *EventHeaderData, seqDepth idx.Event) bool
}

// Block is a "chain" block.
type Block struct {
	Index      idx.Block
	Time       Timestamp
	TxHash     common.Hash
	Events     hash.Events
	SkippedTxs []uint // indexes of skipped txs, starting from first tx of first event, ending with last tx of last event
	GasUsed    uint64

	PrevHash hash.Event

	Root    common.Hash
	Atropos hash.Event
}

// NewBlock makes block from topological ordered events.
func NewBlock(index idx.Block, time Timestamp, atropos hash.Event, prevHash hash.Event, events hash.Events) *Block {
	return &Block{
		Index:      index,
		Time:       time,
		Events:     events,
		PrevHash:   prevHash,
		SkippedTxs: make([]uint, 0),
		Atropos:    atropos,
	}
}
