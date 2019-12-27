package gossip

/*
	In LRU cache data stored like pointer
*/

import (
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
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
		return
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
		return nil
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

	s.tmpDbSet(es.Headers, key, e)

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

	w, _ := s.tmpDbGet(es.Headers, key, &inter.EventHeaderData{}).(*inter.EventHeaderData)

	// Save to LRU cache.
	if w != nil && s.cache.EventsHeaders != nil {
		s.cache.EventsHeaders.Add(h, w)
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
		return
	}

	// Remove from LRU cache.
	if s.cache.EventsHeaders != nil {
		s.cache.EventsHeaders.Remove(h)
	}
}

// tmpDbSet RLP value, ignore if DB is closed
func (s *Store) tmpDbSet(table kvdb.KeyValueStore, key []byte, val interface{}) {
	buf, err := rlp.EncodeToBytes(val)
	if err != nil {
		s.Log.Crit("Failed to encode rlp", "err", err)
	}

	if err := table.Put(key, buf); err != nil {
		return
	}
}

// tmpDbGet RLP value, return nil if DB is closed
func (s *Store) tmpDbGet(table kvdb.KeyValueStore, key []byte, to interface{}) interface{} {
	buf, err := table.Get(key)
	if err != nil {
		return nil // return nil if DB is closed
	}
	if buf == nil {
		return nil
	}

	err = rlp.DecodeBytes(buf, to)
	if err != nil {
		s.Log.Crit("Failed to decode rlp", "err", err, "size", len(buf))
	}
	return to
}
