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
	// LastSuperFrame returns super-frame and list of peers
	LastSuperFrame() (idx.SuperFrame, []hash.Peer)
	// SuperFrame returns list of peers for n super-frame
	SuperFrame(n idx.SuperFrame) []hash.Peer
}
