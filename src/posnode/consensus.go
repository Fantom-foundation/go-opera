package posnode

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/wire"
)

// Consensus is a consensus interface.
type Consensus interface {

	// PushEvent takes event for processing.
	PushEvent(e *wire.Event)

	// GetEvent returns an event by creator and index.
	GetEvent(creator common.Address, index uint64) *wire.Event

	// LastKnownEvent returns an index of last known creators event.
	LastKnownEvent(creator common.Address) uint64
}
