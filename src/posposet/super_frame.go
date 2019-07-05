package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/election"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/seeing"
)

type superFrame struct {
	// state
	frames  map[idx.Frame]*Frame
	members internal.Members

	// election votes
	election *election.Election

	seeing *Strongly
}

func newSuperFrame() *superFrame {
	return &superFrame{
		frames:   make(map[idx.Frame]*Frame),
		members:  internal.Members{},
		election: nil,
	}
}
