package posnode

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

// SetEvent stores event.
func (s *Store) SetEvent(e *inter.Event) {
	s.set(s.Events, e.Hash().Bytes(), e.ToWire())
}

// GetEvent returns stored event.
func (s *Store) GetEvent(h hash.Event) *inter.Event {
	w := s.GetWireEvent(h)
	return inter.WireToEvent(w)
}

// HasEvent returns true if event exists.
func (s *Store) HasEvent(h hash.Event) bool {
	return s.has(s.Events, h.Bytes())
}

// GetWireEvent returns stored event.
// Result is a ready gRPC message.
func (s *Store) GetWireEvent(h hash.Event) *wire.Event {
	w, _ := s.get(s.Events, h.Bytes(), &wire.Event{}).(*wire.Event)
	return w
}

// SetEventHash stores hash.
func (s *Store) SetEventHash(creator hash.Peer, index uint64, hash hash.Event) {
	key := append(creator.Bytes(), intToBytes(index)...)

	if err := s.Hashes.Put(key, hash.Bytes()); err != nil {
		s.Fatal(err)
	}
}

// GetEventHash returns stored event hash.
func (s *Store) GetEventHash(creator hash.Peer, index uint64) *hash.Event {
	key := append(creator.Bytes(), intToBytes(index)...)

	buf, err := s.Hashes.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return nil
	}

	e := hash.BytesToEventHash(buf)
	return &e
}
