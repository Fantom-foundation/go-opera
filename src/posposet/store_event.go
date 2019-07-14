package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// SetEventConfirmedBy stores confirmed event hash.
func (s *Store) SetEventConfirmedBy(e, by hash.Event) {
	key := e.Bytes()

	if err := s.table.ConfirmedEvent.Put(key, by.Bytes()); err != nil {
		s.Fatal(err)
	}
}

// GetEventConfirmedBy returns confirmed event hash.
func (s *Store) GetEventConfirmedBy(e hash.Event) hash.Event {
	key := e.Bytes()

	buf, err := s.table.ConfirmedEvent.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return hash.ZeroEvent
	}

	return hash.BytesToEvent(buf)
}
