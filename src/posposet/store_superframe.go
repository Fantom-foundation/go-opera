package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// SetSuperFrame stores super-frame.
func (s *Store) SetSuperFrame(n idx.SuperFrame, sf *superFrame) {
	s.set(s.table.SuperFrames, n.Bytes(), sf)
}

// GetSuperFrame returns stored super-frame.
func (s *Store) GetSuperFrame(n idx.SuperFrame) *superFrame {
	w, exists := s.get(s.table.SuperFrames, n.Bytes(), &superFrame{}).(*superFrame)
	if !exists {
		return nil
	}
	return w
}
