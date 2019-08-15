package gossip

//go:generate mockgen -package=posnode -self_package=github.com/Fantom-foundation/go-lachesis/src/gossip -destination=mock_consensus.go github.com/Fantom-foundation/go-lachesis/src/gossip Consensus

import (
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
	// StakeOf returns stake of peer.
	StakeOf(hash.Peer) pos.Stake
	// GetGenesisHash returns hash of genesis poset works with.
	GetGenesisHash() hash.Hash
	// GetVectorIndex returns internal vector clock if exists
	GetVectorIndex() *vector.Index
	// Sets consensus fields. Returns nil if event should be dropped.
	Prepare(e *inter.Event) *inter.Event
	// CurrentSuperFrame returns current SuperFrameN.
	CurrentSuperFrameN() idx.SuperFrame
	// SuperFrameMembers returns members of current super-frame.
	SuperFrameMembers() []hash.Peer
}
