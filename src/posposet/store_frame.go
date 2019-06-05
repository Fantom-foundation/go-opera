package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// SetFrame stores event.
func (s *Store) SetFrame(f *Frame) {
	w := f.ToWire()
	s.set(s.Frames, intToBytes(f.Index), w)

	if s.framesCache != nil {
		s.framesCache.Add(f.Index, w)
	}
}

// GetFrame returns stored frame.
func (s *Store) GetFrame(n uint64) *Frame {
	if s.framesCache != nil {
		if f, ok := s.framesCache.Get(n); ok {
			w := f.(*wire.Frame)
			return WireToFrame(w)
		}
	}

	w, _ := s.get(s.Frames, intToBytes(n), &wire.Frame{}).(*wire.Frame)
	return WireToFrame(w)
}

// SetEventFrame stores frame num of event.
func (s *Store) SetEventFrame(e hash.Event, frame uint64) {
	key := e.Bytes()
	val := intToBytes(frame)
	if err := s.Event2frame.Put(key, val); err != nil {
		s.Fatal(err)
	}

	if s.event2frameCache != nil {
		s.event2frameCache.Add(e, frame)
	}
}

// GetEventFrame returns frame num of event.
func (s *Store) GetEventFrame(e hash.Event) *uint64 {
	if s.event2frameCache != nil {
		if n, ok := s.event2frameCache.Get(e); ok {
			num := n.(uint64)
			return &num
		}
	}

	key := e.Bytes()
	buf, err := s.Event2frame.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return nil
	}

	val := bytesToInt(buf)
	return &val
}
