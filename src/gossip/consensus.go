package gossip

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/vector"
)

// Consensus is a consensus interface.
type Consensus interface {
	// PushEvent takes event for processing.
	ProcessEvent(e *inter.Event) error
	// GetGenesisHash returns hash of genesis poset works with.
	GetGenesisHash() common.Hash
	// GetVectorIndex returns internal vector clock if exists
	GetVectorIndex() *vector.Index
	// Sets consensus fields. Returns nil if event should be dropped.
	Prepare(e *inter.Event) *inter.Event
	// LastBlock returns current block.
	LastBlock() (idx.Block, hash.Event)
	// GetEpoch returns current epoch num.
	GetEpoch() idx.Epoch
	// GetValidators returns validators of current epoch.
	GetValidators() pos.Validators
	// GetEpochValidators atomically returns validators of current epoch, and the epoch.
	GetEpochValidators() (pos.Validators, idx.Epoch)
	// GetConsensusTime calc consensus timestamp for given event.
	GetConsensusTime(id hash.Event) (inter.Timestamp, error)

	// Bootstrap must be called (once) before calling other methods
	Bootstrap(applyBlock inter.ApplyBlockFn)
}
