package gossip

/*
	In LRU cache data stored like pointer
*/

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
)

type (
	epochStore struct {
		Headers kvdb.KeyValueStore `table:"header_"`
		Tips    kvdb.KeyValueStore `table:"tips_"`
		Heads   kvdb.KeyValueStore `table:"heads_"`
	}
)

// getEpochStore is not safe for concurrent use.
func (s *Store) getEpochStore(epoch idx.Epoch) *epochStore {
	tables := s.getTmpDb("epoch", uint64(epoch), func(db kvdb.KeyValueStore) interface{} {
		es := &epochStore{}
		table.MigrateTables(es, db)
		return es
	})
	if tables == nil {
		return nil
	}

	return tables.(*epochStore)
}

// delEpochStore is not safe for concurrent use.
func (s *Store) delEpochStore(epoch idx.Epoch) {
	s.delTmpDb("epoch", uint64(epoch))

	// Clear full LRU cache.
	if s.cache.EventsHeaders != nil {
		s.cache.EventsHeaders.Purge()
	}
}

func (s *Store) SetLastEvent(epoch idx.Epoch, from common.Address, id hash.Event) {
	es := s.getEpochStore(epoch)
	if es == nil {
		return
	}

	key := from.Bytes()
	if err := es.Tips.Put(key, id.Bytes()); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (s *Store) GetLastEvent(epoch idx.Epoch, from common.Address) *hash.Event {
	es := s.getEpochStore(epoch)
	if es == nil {
		return nil
	}

	key := from.Bytes()
	idBytes, err := es.Tips.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if idBytes == nil {
		return nil
	}
	id := hash.BytesToEvent(idBytes)
	return &id
}

// SetEventHeader returns stored event header.
func (s *Store) SetEventHeader(epoch idx.Epoch, h hash.Event, e *inter.EventHeaderData) {
	es := s.getEpochStore(epoch)
	if es == nil {
		return
	}

	key := h.Bytes()

	s.set(es.Headers, key, e)

	// Save to LRU cache.
	if s.cache.EventsHeaders != nil {
		s.cache.EventsHeaders.Add(string(key), e)
	}
}

// GetEventHeader returns stored event header.
func (s *Store) GetEventHeader(epoch idx.Epoch, h hash.Event) *inter.EventHeaderData {
	key := h.Bytes()

	// Check LRU cache first.
	if s.cache.EventsHeaders != nil {
		if v, ok := s.cache.EventsHeaders.Get(string(key)); ok {
			if w, ok := v.(*inter.EventHeaderData); ok {
				return w
			}
		}
	}

	es := s.getEpochStore(epoch)
	if es == nil {
		return nil
	}

	w, _ := s.get(es.Headers, key, &inter.EventHeaderData{}).(*inter.EventHeaderData)

	// Save to LRU cache.
	if w != nil && s.cache.EventsHeaders != nil {
		s.cache.EventsHeaders.Add(string(key), w)
	}

	return w
}

// DelEventHeader removes stored event header.
func (s *Store) DelEventHeader(epoch idx.Epoch, h hash.Event) {
	es := s.getEpochStore(epoch)
	if es == nil {
		return
	}

	key := h.Bytes()
	err := es.Headers.Delete(key)
	if err != nil {
		s.Log.Crit("Failed to delete key", "err", err)
	}

	// Remove from LRU cache.
	if s.cache.EventsHeaders != nil {
		s.cache.EventsHeaders.Remove(string(key))
	}
}
