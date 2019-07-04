package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/members"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// SetMembers stores members of super-frame.
func (s *Store) SetMembers(n uint64, mm members.Members) {
	key := common.IntToBytes(n)

	s.set(s.table.Members, key, mm.ToWire())
}

// GetMembers returns stored members of super-frame.
func (s *Store) GetMembers(n uint64) members.Members {
	key := common.IntToBytes(n)

	w := s.get(s.table.Members, key, &wire.Members{}).(*wire.Members)
	return members.WireToMembers(w)
}
