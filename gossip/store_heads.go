package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

func (s *Store) DelHead(epoch idx.Epoch, id hash.Event) {
	es := s.getEpochStore(epoch)
	if es == nil {
		return
	}

	key := id.Bytes()

	if err := es.table.Heads.Delete(key); err != nil {
		s.Log.Crit("Failed to delete key", "err", err)
	}
}

func (s *Store) AddHead(epoch idx.Epoch, id hash.Event) {
	es := s.getEpochStore(epoch)
	if es == nil {
		return
	}

	key := id.Bytes()

	if err := es.table.Heads.Put(key, []byte{}); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (s *Store) IsHead(epoch idx.Epoch, id hash.Event) bool {
	es := s.getEpochStore(epoch)
	if es == nil {
		return false
	}

	key := id.Bytes()

	ok, err := es.table.Heads.Has(key)
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	return ok
}

// GetHeads returns IDs of all the epoch events with no descendants
func (s *Store) GetHeads(epoch idx.Epoch) hash.Events {
	es := s.getEpochStore(epoch)
	if es == nil {
		return nil
	}

	res := make(hash.Events, 0, 100)

	it := es.table.Heads.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		res.Add(hash.BytesToEvent(it.Key()))
	}
	if it.Error() != nil {
		s.Log.Crit("Failed to iterate keys", "err", it.Error())
	}

	return res
}
