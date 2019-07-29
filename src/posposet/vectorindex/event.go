package vectorindex

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

type LowestAfter struct {
	Seq idx.Event
}

type HighestBefore struct {
	Seq         idx.Event
	Id          hash.Event
	ClaimedTime inter.Timestamp
}

// Event is a seeing event for internal purpose.
type Event struct {
	*inter.Event // TODO should be EventHeader

	MemberIdx idx.Member // creator index
	// by node indexes (event idx starts from 1, so 0 means "no one" or "fork"):
	LowestAfter   []LowestAfter   // first events heights who sees it
	HighestBefore []HighestBefore // last events heights who is seen by it
}
