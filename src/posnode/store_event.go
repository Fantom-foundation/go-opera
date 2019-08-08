package posnode

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

// SetEvent stores event.
func (s *Store) SetEvent(e *inter.Event) {
	key := e.Hash().Bytes()

	s.set_rlp(s.table.Events, key, e)
}

// GetEvent returns stored event.
func (s *Store) GetEvent(h hash.Event) *inter.Event {
	key := h.Bytes()

	w, _ := s.get_rlp(s.table.Events, key, &inter.Event{}).(*inter.Event)
	return w
}

// TODO store separately
// GetEventHeader returns stored event header.
func (s *Store) GetEventHeader(h hash.Event) *inter.EventHeaderData {
	e := s.GetEvent(h)
	if e == nil {
		return nil
	}
	return &e.EventHeaderData
}

// HasEvent returns true if event exists.
func (s *Store) HasEvent(h hash.Event) bool {
	return s.has(s.table.Events, h.Bytes())
}

func (s *Store) GetWireEvent(h hash.Event) *wire.Event {
	e := s.GetEvent(h)
	if e == nil {
		return nil
	}

	return e.ToWire()
}

// SetEventHash stores hash.
func (s *Store) SetEventHash(creator hash.Peer, sf idx.SuperFrame, seq idx.Event, hash hash.Event) {

	key := append(creator.Bytes(), sf.Bytes()...)
	key = append(key, seq.Bytes()...)

	if err := s.table.Hashes.Put(key, hash.Bytes()); err != nil {
		s.Fatal(err)
	}
}

// GetEventHash returns stored event hash.
func (s *Store) GetEventHash(creator hash.Peer, sf idx.SuperFrame, seq idx.Event) *hash.Event {
	key := append(creator.Bytes(), sf.Bytes()...)
	key = append(key, seq.Bytes()...)

	buf, err := s.table.Hashes.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return nil
	}

	e := hash.BytesToEvent(buf)
	return &e
}
