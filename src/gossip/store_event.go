package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// TODO store separately
// GetEventHeader returns stored event header.
func (s *Store) GetEventHeader(h hash.Event) *inter.EventHeaderData {
	e := s.GetEvent(h)
	if e == nil {
		return nil
	}
	return &e.EventHeaderData
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
func (s *Store) GetEvent(h hash.Event) *inter.Event {
	key := h.Bytes()

	w, _ := s.get(s.table.Events, key, &inter.Event{}).(*inter.Event)
	return w
}

// HasEvent returns true if event exists.
func (s *Store) HasEvent(h hash.Event) bool {
	return s.has(s.table.Events, h.Bytes())
}
