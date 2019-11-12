package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/ethereum/go-ethereum/common"
)

func (s *Store) deleteLastHeaders() {
	prevKeys := make([][]byte, 0, 500) // don't write during iteration

	it := s.table.LastEpochHeaders.NewIterator()
	defer it.Release()
	for it.Next() {
		prevKeys = append(prevKeys, it.Key())
	}
	for _, key := range prevKeys {
		err := s.table.LastEpochHeaders.Delete(key)
		if err != nil {
			s.Log.Crit("Failed to erase key-value", "err", err)
		}
	}
	// Add to cache.
	s.cache.LastEpochHeaders = nil
}

func (s *Store) SetLastHeaders(hh inter.HeadersByCreator) {
	// delete previous headers
	s.deleteLastHeaders()
	// put new headers
	for creator, header := range hh {
		err := s.table.LastEpochHeaders.Put(creator.Bytes(), header.Hash().Bytes())
		if err != nil {
			s.Log.Crit("Failed to put key-value", "err", err)
		}
	}
	// Add to cache.
	s.cache.LastEpochHeaders = hh
}

func (s *Store) GetLastHeaders() inter.HeadersByCreator {
	// Get LastHeaders from cache first.
	if s.cache.LastEpochHeaders != nil {
		return s.cache.LastEpochHeaders
	}

	hh := make(inter.HeadersByCreator)

	it := s.table.LastEpochHeaders.NewIterator()
	defer it.Release()
	for it.Next() {
		id := hash.BytesToEvent(it.Value())
		header := s.GetEventHeader(id.Epoch(), id)
		if header == nil {
			s.Log.Crit("Failed to load LastHeaders", "err", "not found")
		}
		hh[common.BytesToAddress(it.Key())] = header
	}

	// Add to cache.
	s.cache.LastEpochHeaders = hh

	return hh
}

func (s *Store) AddDirtyLastHeader(epoch idx.Epoch, creator common.Address, id hash.Event) {
	es := s.getEpochStore(epoch)
	if es == nil {
		return
	}

	err := es.LastDirtyEpochHeaders.Put(creator.Bytes(), id.Bytes())
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (s *Store) GetDirtyLastHeaders(epoch idx.Epoch) inter.HeadersByCreator {
	es := s.getEpochStore(epoch)
	if es == nil {
		return nil
	}

	hh := make(inter.HeadersByCreator)

	it := es.LastDirtyEpochHeaders.NewIterator()
	defer it.Release()
	for it.Next() {
		id := hash.BytesToEvent(it.Value())
		header := s.GetEventHeader(id.Epoch(), id)
		if header == nil {
			s.Log.Crit("Failed to load LastHeaders", "err", "not found")
		}
		hh[common.BytesToAddress(it.Key())] = header
	}
	return hh
}
