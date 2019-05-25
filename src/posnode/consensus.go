package posnode

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// Consensus is a consensus interface.
type Consensus interface {
	// PushEvent takes event for processing.
	PushEvent(hash.Event)
	// GetStakeOf returns stake of peer as fraction from one.
	GetStakeOf(hash.Peer) float64
	// GetTransactionCount returns cound of transactions made
	// by peer.
	GetTransactionCount(hash.Peer) uint64
}
