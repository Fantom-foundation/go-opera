package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

func (s *Store) EraseHead(id hash.Event) {
	key := id.Bytes()

	if err := s.table.Heads.Delete(key); err != nil {
		s.Fatal(err)
	}
}

func (s *Store) AddHead(id hash.Event) {
	key := id.Bytes()

	if err := s.table.Heads.Put(key, []byte{}); err != nil {
		s.Fatal(err)
	}
}

func (s *Store) IsHead(id hash.Event) bool {
	key := id.Bytes()

	ok, err := s.table.Heads.Has(key)
	if err != nil {
		s.Fatal(err)
	}
	return ok
}

// GetHeads returns all the events with no descendants
func (s *Store) GetHeads() hash.Events {
	prefix := []byte{}
	res := []hash.Event{}
	err := s.table.Heads.ForEach(prefix, func(key, _ []byte) bool {
		res = append(res, hash.BytesToEvent(key))
		return true
	})
	if err != nil {
		s.Fatal(err)
	}
	return res
}
