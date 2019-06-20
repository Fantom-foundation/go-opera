package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// SetFrame stores event.
func (s *Store) SetFrame(f *Frame) {
	key := common.IntToBytes(f.Index)
	w := f.ToWire()
	s.set(s.table.Frames, key, w)

	if s.cache.Frames != nil {
		s.cache.Frames.Add(f.Index, w)

		// for metrics
		updateCacheInfo(
			s.cache.Frames,
			frameCacheSize,
			frameCacheCap,
			w,
		)
	}
}

// GetFrame returns stored frame.
func (s *Store) GetFrame(n uint64) *Frame {
	if s.cache.Frames != nil {
		if f, ok := s.cache.Frames.Get(n); ok {
			w := f.(*wire.Frame)
			return WireToFrame(w)
		}
	}

	key := common.IntToBytes(n)
	w, _ := s.get(s.table.Frames, key, &wire.Frame{}).(*wire.Frame)
	return WireToFrame(w)
}

// SetEventFrame stores frame num of event.
func (s *Store) SetEventFrame(e hash.Event, frame uint64) {
	key := e.Bytes()
	val := common.IntToBytes(frame)
	if err := s.table.Event2Frame.Put(key, val); err != nil {
		s.Fatal(err)
	}

	if s.cache.Event2Frame != nil {
		s.cache.Event2Frame.Add(e, frame)

		// for metrics
		updateCacheInfo(
			s.cache.Event2Frame,
			event2FrameCacheCap,
			event2FrameCacheSize,
			frame,
		)
	}
}

// GetEventFrame returns frame num of event.
func (s *Store) GetEventFrame(e hash.Event) *uint64 {
	if s.cache.Event2Frame != nil {
		if n, ok := s.cache.Event2Frame.Get(e); ok {
			num := n.(uint64)
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

	val := common.BytesToInt(buf)
	return &val
}
