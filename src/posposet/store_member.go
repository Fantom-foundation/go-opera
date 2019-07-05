package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// SetMembers stores members of super-frame.
func (s *Store) SetMembers(n idx.SuperFrame, mm internal.Members) {
	s.set(s.table.Members, n.Bytes(), mm.ToWire())
}

// GetMembers returns stored members of super-frame.
func (s *Store) GetMembers(n idx.SuperFrame) internal.Members {
	w := s.get(s.table.Members, n.Bytes(), &wire.Members{}).(*wire.Members)
	return internal.WireToMembers(w)
}
