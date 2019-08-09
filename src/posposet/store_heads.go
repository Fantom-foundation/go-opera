package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

func (s *Store) EraseHead(id hash.Event) {
	key := id.Bytes()

	if err := s.epochTable.Heads.Delete(key); err != nil {
		s.Fatal(err)
	}
}

func (s *Store) AddHead(id hash.Event) {
	key := id.Bytes()

	if err := s.epochTable.Heads.Put(key, []byte{}); err != nil {
		s.Fatal(err)
	}
}

func (s *Store) IsHead(id hash.Event) bool {
	key := id.Bytes()

	ok, err := s.epochTable.Heads.Has(key)
	if err != nil {
		s.Fatal(err)
	}
	return ok
}

func (s *Store) GetHeads() []hash.Event {
	prefix := []byte{}
	res := []hash.Event{}
	err := s.epochTable.Heads.ForEach(prefix, func(key, _ []byte) bool {
		res = append(res, hash.BytesToEvent(key))
		return true
	})
	if err != nil {
		s.Fatal(err)
	}
	return res
}
