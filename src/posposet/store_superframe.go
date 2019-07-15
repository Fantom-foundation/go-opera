package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// SetSuperFrame stores super-frame.
func (s *Store) SetSuperFrame(n idx.SuperFrame, sf *superFrame) {
	s.set(s.table.SuperFrames, n.Bytes(), sf.ToWire())
}

// GetMembers returns stored super-frame.
func (s *Store) GetSuperFrame(n idx.SuperFrame) *superFrame {
	w := s.get(s.table.SuperFrames, n.Bytes(), &wire.SuperFrame{}).(*wire.SuperFrame)
	return WireToSuperFrame(w)
}
