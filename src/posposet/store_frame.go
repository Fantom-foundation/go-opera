package posposet

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// SetFrame stores event.
func (s *Store) SetFrame(f *Frame, sf idx.SuperFrame) {
	key := []byte(fmt.Sprintf("%d_%d", sf, f.Index))

	w := f.ToWire()
	s.set(s.table.Frames, key, w)

	if s.cache.Frames != nil {
		s.cache.Frames.Add(f.Index, w)

		frameCacheCap.Update(int64(
			s.cache.Frames.Len()))
	}
}

// GetFrame returns stored frame.
func (s *Store) GetFrame(n idx.Frame, sf idx.SuperFrame) *Frame {
	key := []byte(fmt.Sprintf("%d_%d", sf, n))

	if s.cache.Frames != nil {
		if f, ok := s.cache.Frames.Get(n); ok {
			w := f.(*wire.Frame)
			return WireToFrame(w)
		}
	}

	w, _ := s.get(s.table.Frames, key, &wire.Frame{}).(*wire.Frame)
	return WireToFrame(w)
}

// SetEventFrame stores frame num of event.
func (s *Store) SetEventFrame(e hash.Event, frame idx.Frame) {
	key := e.Bytes()
	val := frame.Bytes()
	if err := s.table.Event2Frame.Put(key, val); err != nil {
		s.Fatal(err)
	}

	if s.cache.Event2Frame != nil {
		s.cache.Event2Frame.Add(e, frame)

		event2FrameCacheCap.Update(int64(
			s.cache.Event2Frame.Len()))
	}
}

// GetEventFrame returns frame num of event.
func (s *Store) GetEventFrame(e hash.Event) *idx.Frame {
	if s.cache.Event2Frame != nil {
		if n, ok := s.cache.Event2Frame.Get(e); ok {
			num := n.(idx.Frame)
			return &num
		}
	}

	key := e.Bytes()
	buf, err := s.table.Event2Frame.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return nil
	}

	val := idx.BytesToFrame(buf)
	return &val
}
