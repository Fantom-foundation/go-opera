package poset

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// SetFrameInfo stores event.
func (s *Store) SetFrameInfo(e idx.Epoch, f idx.Frame, info *FrameInfo) {
	key := fmt.Sprintf("%d_%d", e, f)

	s.set(s.table.FrameInfos, []byte(key), info)
}

// GetFrameInfo returns stored frame.
func (s *Store) GetFrameInfo(e idx.Epoch, f idx.Frame) *FrameInfo {
	key := fmt.Sprintf("%d_%d", e, f)

	w, exists := s.get(s.table.FrameInfos, []byte(key), &FrameInfo{}).(*FrameInfo)
	if !exists {
		return nil
	}

	return w
}
