package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/ethereum/go-ethereum/rlp"
)

// DeleteEvent deletes event.
func (s *Store) DeleteEvent(id hash.Event) {
	key := id.Bytes()

	err := s.table.Events.Delete(key)
	if err != nil {
		s.Fatal(err)
	}
	err = s.table.Headers.Delete(key)
	if err != nil {
		s.Fatal(err)
	}
}

// SetEvent stores event.
func (s *Store) SetEvent(e *inter.Event) {
	key := e.Hash().Bytes()

	s.set(s.table.Events, key, e)
	s.set(s.table.Headers, key, e.EventHeaderData)
}

// GetEventHeader returns stored event header.
func (s *Store) GetEventHeader(h hash.Event) *inter.EventHeaderData {
	key := h.Bytes()

	w, _ := s.get(s.table.Headers, key, &inter.EventHeaderData{}).(*inter.EventHeaderData)
	return w
}

// GetEvent returns stored event.
func (s *Store) GetEvent(id hash.Event) *inter.Event {
	key := id.Bytes()

	w, _ := s.get(s.table.Events, key, &inter.Event{}).(*inter.Event)
	return w
}

// GetEventRLP returns stored event. Serialized.
func (s *Store) GetEventRLP(id hash.Event) rlp.RawValue {
	key := id.Bytes()

	data, err := s.table.Events.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	return data
}

// HasEvent returns true if event exists.
func (s *Store) HasEvent(h hash.Event) bool {
	return s.has(s.table.Events, h.Bytes())
}
