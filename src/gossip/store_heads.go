package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

func (s *Store) DelHead(epoch idx.SuperFrame, id hash.Event) {
	es := s.getEpochStore(epoch)
	if es == nil {
		return
	}

	key := id.Bytes()

	if err := es.Heads.Delete(key); err != nil {
		s.Fatal(err)
	}
}

func (s *Store) AddHead(epoch idx.SuperFrame, id hash.Event) {
	es := s.getEpochStore(epoch)
	if es == nil {
		return
	}

	key := id.Bytes()

	if err := es.Heads.Put(key, []byte{}); err != nil {
		s.Fatal(err)
	}
}

func (s *Store) IsHead(epoch idx.SuperFrame, id hash.Event) bool {
	es := s.getEpochStore(epoch)
	if es == nil {
		return false
	}

	key := id.Bytes()

	ok, err := es.Heads.Has(key)
	if err != nil {
		s.Fatal(err)
	}
	return ok
}

// GetHeads returns all the events with no descendants.
func (s *Store) GetHeads(epoch idx.SuperFrame) hash.Events {
	es := s.getEpochStore(epoch)
	if es == nil {
		return nil
	}

	prefix := []byte{}
	res := []hash.Event{}
	err := es.Heads.ForEach(prefix, func(key, _ []byte) bool {
		res = append(res, hash.BytesToEvent(key))
		return true
	})
	if err != nil {
		s.Fatal(err)
	}
	return res
}
