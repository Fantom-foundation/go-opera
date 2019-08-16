package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// SetEventConfirmedOn stores confirmed event hash.
func (s *Store) SetEventConfirmedOn(e hash.Event, on idx.Frame) {
	key := e.Bytes()

	if err := s.table.ConfirmedEvent.Put(key, on.Bytes()); err != nil {
		s.Fatal(err)
	}
}

// GetEventConfirmedOn returns confirmed event hash.
func (s *Store) GetEventConfirmedOn(e hash.Event) idx.Frame {
	key := e.Bytes()

	buf, err := s.table.ConfirmedEvent.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return 0
	}

	return idx.BytesToFrame(buf)
}
