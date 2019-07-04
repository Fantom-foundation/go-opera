package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/posposet/election"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/members"
)

type superFrame struct {
	// state
	frames  map[uint64]*Frame
	members members.Members

	// election votes
	election *election.Election
}

func newSuperFrame() *superFrame {
	return &superFrame{
		frames:   make(map[uint64]*Frame),
		members:  members.Members{},
		election: nil,
	}
}
