package posposet

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// SetFrame stores event.
func (s *Store) SetFrame(sf idx.SuperFrame, f *Frame) {
	key := fmt.Sprintf("%d_%d", sf, f.Index)

	w := f.ToWire()
	s.set(s.table.Frames, []byte(key), w)

	if s.cache.Frames != nil {
		s.cache.Frames.Add(key, w)

		frameCacheCap.Update(int64(
			s.cache.Frames.Len()))
	}
}

// GetFrame returns stored frame.
func (s *Store) GetFrame(sf idx.SuperFrame, n idx.Frame) *Frame {
	key := fmt.Sprintf("%d_%d", sf, n)

	if s.cache.Frames != nil {
		if f, ok := s.cache.Frames.Get(key); ok {
			w := f.(*wire.Frame)
			return WireToFrame(w)
		}
	}

	w, _ := s.get(s.table.Frames, []byte(key), &wire.Frame{}).(*wire.Frame)
	return WireToFrame(w)
}
