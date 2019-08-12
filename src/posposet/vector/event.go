package vector

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

type (
	lowestAfter struct {
		Seq idx.Event
	}

	highestBefore struct {
		IsForkSeen  bool
		Seq         idx.Event
		ID          hash.Event
		ClaimedTime inter.Timestamp
	}

	// event is an inter.EventHeaderData wrapper for internal purpose.
	event struct {
		*inter.EventHeaderData

		CreatorIdx idx.Member // creator index
		// by node indexes (event idx starts from 1, so 0 means "no one" or "fork"):
		LowestAfter   []lowestAfter   // first events heights who sees it
		HighestBefore []highestBefore // last events heights who is seen by it
	}
)
