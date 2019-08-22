package gossip

import (
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// DeleteEvent deletes event.
func (s *Store) DeleteEvent(epoch idx.SuperFrame, id hash.Event) {
	key := id.Bytes()

	err := s.table.Events.Delete(key)
	if err != nil {
		s.Fatal(err)
	}
	s.DelEventHeader(epoch, id)
}

// SetEvent stores event.
func (s *Store) SetEvent(e *inter.Event) {
	key := e.Hash().Bytes()

	s.set(s.table.Events, key, e)
	s.SetEventHeader(e.Epoch, e.Hash(), &e.EventHeaderData)
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
