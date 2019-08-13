//go:generate mockgen -package=posnode -self_package=github.com/Fantom-foundation/go-lachesis/src/posnode -destination=mock_consensus.go github.com/Fantom-foundation/go-lachesis/src/posnode Consensus
package posnode

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// Consensus is a consensus interface.
type Consensus interface {
	// PushEvent takes event for processing.
	PushEvent(hash.Event)
	// StakeOf returns stake of peer.
	StakeOf(hash.Peer) inter.Stake
	// GetGenesisHash returns hash of genesis poset works with.
	GetGenesisHash() hash.Hash
	// Sets consensus fields. Returns nil if event should be dropped.
	Prepare(e *inter.Event) *inter.Event
	// CurrentSuperFrame returns current SuperFrameN.
	CurrentSuperFrameN() idx.SuperFrame
	// SuperFrameMembers returns members of current super-frame.
	SuperFrameMembers() []hash.Peer
}
