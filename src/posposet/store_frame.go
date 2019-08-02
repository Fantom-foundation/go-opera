package posposet

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// SetFrame stores event.
func (s *Store) SetFrame(sf idx.SuperFrame, f *Frame) {
	key := fmt.Sprintf("%d_%d", sf, f.Index)

	s.set(s.table.Frames, []byte(key), f)
}

// GetFrame returns stored frame.
func (s *Store) GetFrame(sf idx.SuperFrame, n idx.Frame) *Frame {
	key := fmt.Sprintf("%d_%d", sf, n)

	w, _ := s.get(s.table.Frames, []byte(key), &Frame{}).(*Frame)
	return w
}
