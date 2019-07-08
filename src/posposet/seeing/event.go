package seeing

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// Event is a seeing event for internal purpose.
type Event struct {
	*inter.Event

	MemberN int // creator index
	// by node indexes (event idx starts from 1, so 0 means "no one"):
	LowestSees  []idx.Event // first events heights who sees it
	HighestSeen []idx.Event // last events heights who is seen by it
}
