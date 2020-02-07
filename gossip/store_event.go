package gossip

/*
	In LRU cache data stored like pointer
*/

import (
	"bytes"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// DeleteEvent deletes event.
func (s *Store) DeleteEvent(epoch idx.Epoch, id hash.Event) {
	key := id.Bytes()

	err := s.table.Events.Delete(key)
	if err != nil {
		s.Log.Crit("Failed to delete key", "err", err)
	}
	s.DelEventHeader(epoch, id)

	// Remove from LRU cache.
	if s.cache.Events != nil {
		s.cache.Events.Remove(id)
	}

	s.Log.Info("DeleteEvent", "event", id)
}

// SetEvent stores event.
func (s *Store) SetEvent(e *inter.Event) {
	key := e.Hash().Bytes()

	s.set(s.table.Events, key, e)
	s.SetEventHeader(e.Epoch, e.Hash(), &e.EventHeaderData)

	// Add to LRU cache.
	if s.cache.Events != nil {
		s.cache.Events.Add(e.Hash(), e)
	}
}

// GetEvent returns stored event.
func (s *Store) GetEvent(id hash.Event) *inter.Event {
	key := id.Bytes()

	// Get event from LRU cache first.
	if s.cache.Events != nil {
		if c, ok := s.cache.Events.Get(id); ok {
			if ev, ok := c.(*inter.Event); ok {
				return ev
			}
		}
	}

	w, _ := s.get(s.table.Events, key, &inter.Event{}).(*inter.Event)

	// Put event to LRU cache.
	if w != nil && s.cache.Events != nil {
		s.cache.Events.Add(id, w)
	}

	return w
}

func (s *Store) ForEachEvent(epoch idx.Epoch, onEvent func(event *inter.Event) bool) {
	it := s.table.Events.NewIteratorWithPrefix(epoch.Bytes())
	defer it.Release()
	for it.Next() {
		event := &inter.Event{}
		err := rlp.DecodeBytes(it.Value(), event)
		if err != nil {
			s.Log.Crit("Failed to decode event", "err", err)
		}

		if !onEvent(event) {
			return
		}
	}
}

func (s *Store) FindEventHashes(epoch idx.Epoch, lamport idx.Lamport, hashPrefix []byte) hash.Events {
	prefix := bytes.NewBuffer(epoch.Bytes())
	prefix.Write(lamport.Bytes())
	prefix.Write(hashPrefix)
	res := make(hash.Events, 0, 10)

	it := s.table.Events.NewIteratorWithPrefix(prefix.Bytes())
	defer it.Release()
	for it.Next() {
		res = append(res, hash.BytesToEvent(it.Key()))
	}

	return res
}

// GetEventRLP returns stored event. Serialized.
func (s *Store) GetEventRLP(id hash.Event) rlp.RawValue {
	key := id.Bytes()

	data, err := s.table.Events.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	return data
}

// HasEvent returns true if event exists.
func (s *Store) HasEvent(h hash.Event) bool {
	return s.has(s.table.Events, h.Bytes())
}
