package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// SetConfirmedEvent stores confirmed event hash.
func (s *Store) SetConfirmedEvent(h hash.Event, f idx.Frame) {
	key := h.Bytes()
	w := common.IntToBytes(uint64(f))

	if err := s.table.ConfirmedEvent.Put(key, w); err != nil {
		s.Fatal(err)
	}
}

// GetConfirmedEvent returns stored confirmed event hash.
func (s *Store) GetConfirmedEvent(h hash.Event) idx.Frame {
	buf, err := s.table.ConfirmedEvent.Get(h.Bytes())
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return 0
	}

	return idx.BytesToFrame(buf)
}
