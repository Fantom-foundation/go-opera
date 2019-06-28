//go:generate mockgen -package=posnode -self_package=github.com/Fantom-foundation/go-lachesis/src/posnode -destination=mock_consensus.go github.com/Fantom-foundation/go-lachesis/src/posnode Consensus
package posnode

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// Consensus is a consensus interface.
type Consensus interface {
	// PushEvent takes event for processing.
	PushEvent(hash.Event)
	// StakeOf returns stake of peer.
	StakeOf(hash.Peer) uint64
	// GetGenesisHash returns hash of genesis poset works with.
	GetGenesisHash() hash.Hash
}
