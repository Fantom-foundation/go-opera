package poset

import (
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// SetFrameInfo stores event.
func (s *Store) SetFrameInfo(e idx.Epoch, f idx.Frame, info *FrameInfo) {
	key := append(e.Bytes(), f.Bytes()...)

	s.set(s.table.FrameInfos, key, info)
}

// GetFrameInfo returns stored frame.
func (s *Store) GetFrameInfo(e idx.Epoch, f idx.Frame) *FrameInfo {
	key := append(e.Bytes(), f.Bytes()...)

	w, exists := s.get(s.table.FrameInfos, key, &FrameInfo{}).(*FrameInfo)
	if !exists {
		return nil
	}

	return w
}
