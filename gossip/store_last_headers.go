package gossip

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// DelLastHeader deletes record about last header from a validator
func (s *Store) DelLastHeader(epoch idx.Epoch, creator common.Address) {
	s.mutexes.LastEpochHeaders.Lock() // need mutex because of complex mutable cache
	defer s.mutexes.LastEpochHeaders.Unlock()

	key := bytes.NewBuffer(nil)
	key.Write(epoch.Bytes())
	key.Write(creator.Bytes())

	err := s.table.LastEpochHeaders.Delete(key.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase LastHeader", "err", err)
	}

	// Add to cache.
	if s.cache.LastEpochHeaders != nil {
		if c, ok := s.cache.LastEpochHeaders.Get(epoch); ok {
			if hh, ok := c.(inter.HeadersByCreator); ok {
				delete(hh, creator)
			}
		}
	}
}

// DelLastHeaders deletes all the records about last headers
func (s *Store) DelLastHeaders(epoch idx.Epoch) {
	s.mutexes.LastEpochHeaders.Lock() // need mutex because of complex mutable cache
	defer s.mutexes.LastEpochHeaders.Unlock()

	keys := make([][]byte, 0, 500) // don't write during iteration

	it := s.table.LastEpochHeaders.NewIteratorWithPrefix(epoch.Bytes())
	defer it.Release()
	for it.Next() {
		keys = append(keys, it.Key())
	}

	for _, key := range keys {
		err := s.table.LastEpochHeaders.Delete(key)
		if err != nil {
			s.Log.Crit("Failed to erase LastHeader", "err", err)
		}
	}

	// Add to cache.
	if s.cache.LastEpochHeaders != nil {
		s.cache.LastEpochHeaders.Remove(epoch)
	}
}

// AddLastHeader adds/updates a records about last header from a validator
func (s *Store) AddLastHeader(epoch idx.Epoch, header *inter.EventHeaderData) {
	s.mutexes.LastEpochHeaders.Lock() // need mutex because of complex mutable cache
	defer s.mutexes.LastEpochHeaders.Unlock()

	key := bytes.NewBuffer(nil)
	key.Write(epoch.Bytes())
	key.Write(header.Creator.Bytes())

	s.set(s.table.LastEpochHeaders, key.Bytes(), header)

	// Add to cache.
	if s.cache.LastEpochHeaders != nil {
		if c, ok := s.cache.LastEpochHeaders.Get(epoch); ok {
			if hh, ok := c.(inter.HeadersByCreator); ok {
				hh[header.Creator] = header
			}
		}
	}
}

// GetLastHeaders retrieves all the records about last headers from validators
func (s *Store) GetLastHeaders(epoch idx.Epoch) inter.HeadersByCreator {
	s.mutexes.LastEpochHeaders.RLock()
	defer s.mutexes.LastEpochHeaders.RUnlock()

	// Get data from LRU cache first.
	if s.cache.LastEpochHeaders != nil {
		if c, ok := s.cache.LastEpochHeaders.Get(epoch); ok {
			if hh, ok := c.(inter.HeadersByCreator); ok {
				return hh
			}
		}
	}

	hh := make(inter.HeadersByCreator)

	it := s.table.LastEpochHeaders.NewIteratorWithPrefix(epoch.Bytes())
	defer it.Release()
	for it.Next() {
		creator := it.Key()[4:]
		header := &inter.EventHeaderData{}
		err := rlp.DecodeBytes(it.Value(), header)
		if err != nil {
			s.Log.Crit("Failed to decode rlp", "err", err)
		}
		hh[common.BytesToAddress(creator)] = header
	}

	// Add to cache.
	if s.cache.LastEpochHeaders != nil {
		s.cache.LastEpochHeaders.Add(epoch, hh)
	}

	return hh
}
