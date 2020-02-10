package gossip

/*
	In LRU cache data stored like pointer
*/

import (
	"errors"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb"

	"github.com/Fantom-foundation/go-lachesis/kvdb/skiperrors"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
)

type (
	epochStore struct {
		Headers kvdb.KeyValueStore `table:"h"`
		Tips    kvdb.KeyValueStore `table:"t"`
		Heads   kvdb.KeyValueStore `table:"H"`
	}
)

func newEpochStore(db kvdb.KeyValueStore) *epochStore {
	es := &epochStore{}
	table.MigrateTables(es, db)

	err := errors.New("database closed")

	es.Headers = skiperrors.Wrap(es.Headers, err)
	es.Tips = skiperrors.Wrap(es.Tips, err)
	es.Heads = skiperrors.Wrap(es.Heads, err)

	return es
}

// getEpochStore is not safe for concurrent use.
func (s *Store) getEpochStore(epoch idx.Epoch) *epochStore {
	tables := s.EpochDbs.Get(uint64(epoch))
	if tables == nil {
		return nil
	}

	return tables.(*epochStore)
}

// delEpochStore is not safe for concurrent use.
func (s *Store) delEpochStore(epoch idx.Epoch) {
	s.EpochDbs.Del(uint64(epoch))
	// Clear full LRU cache.
	if s.cache.EventsHeaders != nil {
		s.cache.EventsHeaders.Purge()
	}
}

// SetLastEvent stores last unconfirmed event from a validator (off-chain)
func (s *Store) SetLastEvent(epoch idx.Epoch, from idx.StakerID, id hash.Event) {
	es := s.getEpochStore(epoch)
	if es == nil {
		return
	}

	key := from.Bytes()
	if err := es.Tips.Put(key, id.Bytes()); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// GetLastEvent returns stored last unconfirmed event from a validator (off-chain)
func (s *Store) GetLastEvent(epoch idx.Epoch, from idx.StakerID) *hash.Event {
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
	if e != nil && s.cache.EventsHeaders != nil {
		s.cache.EventsHeaders.Add(h, e)
	}
}

// GetEventHeader returns stored event header.
func (s *Store) GetEventHeader(epoch idx.Epoch, h hash.Event) *inter.EventHeaderData {
	key := h.Bytes()

	// Check LRU cache first.
	if s.cache.EventsHeaders != nil {
		if v, ok := s.cache.EventsHeaders.Get(h); ok {
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
		s.cache.EventsHeaders.Add(h, w)
	}

	return w
}

// HasEvent returns true if event exists.
func (s *Store) HasEventHeader(h hash.Event) bool {
	es := s.getEpochStore(h.Epoch())
	if es == nil {
		return false
	}

	return s.has(es.Headers, h.Bytes())
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
		s.cache.EventsHeaders.Remove(h)
	}
}
