package vector

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

type lowestAfter struct {
	Seq idx.Event
}

type highestBefore struct {
	IsForkSeen  bool
	Seq         idx.Event
	ID          hash.Event
	ClaimedTime inter.Timestamp
}

// event is a inter.Event with additional fields for internal purpose.
type event struct {
	*inter.Event // TODO: should be EventHeader

	CreatorIdx idx.Member // creator index
	// by node indexes (event idx starts from 1, so 0 means "no one" or "fork"):
	LowestAfter   []lowestAfter   // first events heights who sees it
	HighestBefore []highestBefore // last events heights who is seen by it
}
