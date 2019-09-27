package poset

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// SetFrameInfo stores event.
func (s *Store) SetFrameInfo(e idx.Epoch, f idx.Frame, info *FrameInfo) {
	key := append(e.Bytes(), f.Bytes()...)

	s.set(s.table.FrameInfos, key, info)
}

// GetFrameInfo returns stored frame.
func (s *Store) GetFrameInfo(e idx.Epoch, f idx.Frame) *FrameInfo {
	key := append(e.Bytes(), f.Bytes()...)

	w, _ := s.get(s.table.FrameInfos, key, &FrameInfo{}).(*FrameInfo)
	return w
}
