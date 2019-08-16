package posposet

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// SetFrameInfo stores event.
func (s *Store) SetFrameInfo(sf idx.SuperFrame, f idx.Frame, info *FrameInfo) {
	key := fmt.Sprintf("%d_%d", sf, f)

	s.set(s.table.FrameInfos, []byte(key), f)
}

// GetFrameInfo returns stored frame.
func (s *Store) GetFrameInfo(sf idx.SuperFrame, n idx.Frame) *FrameInfo {
	key := fmt.Sprintf("%d_%d", sf, n)

	w, _ := s.get(s.table.FrameInfos, []byte(key), &FrameInfo{}).(*FrameInfo)
	return w
}
