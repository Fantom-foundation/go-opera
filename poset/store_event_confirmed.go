package poset

import (
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// SetEventConfirmedOn stores confirmed event hash.
func (s *Store) SetEventConfirmedOn(e hash.Event, on idx.Frame) {
	key := e.Bytes()

	if err := s.table.ConfirmedEvent.Put(key, on.Bytes()); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// GetEventConfirmedOn returns confirmed event hash.
func (s *Store) GetEventConfirmedOn(e hash.Event) idx.Frame {
	key := e.Bytes()

	buf, err := s.table.ConfirmedEvent.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return 0
	}

	return idx.BytesToFrame(buf)
}
