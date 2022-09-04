package gossip

/*
	In LRU cache data stored like pointer
*/

import (
	"bytes"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/inter"
)

// DelEvent deletes event.
func (s *Store) DelEvent(id hash.Event) {
	key := id.Bytes()

	err := s.table.Events.Delete(key)
	if err != nil {
		s.Log.Crit("Failed to delete key", "err", err)
	}

	// Remove from LRU cache.
	s.cache.Events.Remove(id)
	s.cache.EventsHeaders.Remove(id)
	s.cache.EventIDs.Remove(id)
}

// SetEvent stores event.
func (s *Store) SetEvent(e *inter.EventPayload) {
	key := e.ID().Bytes()

	s.rlp.Set(s.table.Events, key, e)

	// Add to LRU cache.
	s.cache.Events.Add(e.ID(), e, uint(e.Size()))
	eh := e.Event
	s.cache.EventsHeaders.Add(e.ID(), &eh, nominalSize)
	s.cache.EventIDs.Add(e.ID())
}

// GetEventPayload returns stored event.
func (s *Store) GetEventPayload(id hash.Event) *inter.EventPayload {
	// Get event from LRU cache first.
	if ev, ok := s.cache.Events.Get(id); ok {
		return ev.(*inter.EventPayload)
	}

	key := id.Bytes()
	w, _ := s.rlp.Get(s.table.Events, key, &inter.EventPayload{}).(*inter.EventPayload)

	if w != nil {
		fixEventTxHashes(w)
	}

	// Put event to LRU cache.
	if w != nil {
		s.cache.Events.Add(id, w, uint(w.Size()))
		eh := w.Event
		s.cache.EventsHeaders.Add(id, &eh, nominalSize)
	}

	return w
}

// GetEvent returns stored event.
func (s *Store) GetEvent(id hash.Event) *inter.Event {
	// Get event from LRU cache first.
	if ev, ok := s.cache.EventsHeaders.Get(id); ok {
		return ev.(*inter.Event)
	}

	key := id.Bytes()
	w, _ := s.rlp.Get(s.table.Events, key, &inter.EventPayload{}).(*inter.EventPayload)
	if w == nil {
		return nil
	}
	fixEventTxHashes(w)

	eh := w.Event

	// Put event to LRU cache.
	s.cache.Events.Add(id, w, uint(w.Size()))
	s.cache.EventsHeaders.Add(id, &eh, nominalSize)

	return &eh
}

func (s *Store) forEachEvent(it ethdb.Iterator, onEvent func(event *inter.EventPayload) bool) {
	for it.Next() {
		event := &inter.EventPayload{}
		err := rlp.DecodeBytes(it.Value(), event)
		if err != nil {
			s.Log.Crit("Failed to decode event", "err", err)
		}

		if !onEvent(event) {
			return
		}
	}
}

func (s *Store) ForEachEpochEvent(epoch idx.Epoch, onEvent func(event *inter.EventPayload) bool) {
	it := s.table.Events.NewIterator(epoch.Bytes(), nil)
	defer it.Release()
	s.forEachEvent(it, onEvent)
}

func (s *Store) ForEachEvent(start idx.Epoch, onEvent func(event *inter.EventPayload) bool) {
	it := s.table.Events.NewIterator(nil, start.Bytes())
	defer it.Release()
	s.forEachEvent(it, onEvent)
}

func (s *Store) ForEachEventRLP(start []byte, onEvent func(key hash.Event, event rlp.RawValue) bool) {
	it := s.table.Events.NewIterator(nil, start)
	defer it.Release()
	for it.Next() {
		if !onEvent(hash.BytesToEvent(it.Key()), it.Value()) {
			return
		}
	}
}

func (s *Store) FindEventHashes(epoch idx.Epoch, lamport idx.Lamport, hashPrefix []byte) hash.Events {
	prefix := bytes.NewBuffer(epoch.Bytes())
	prefix.Write(lamport.Bytes())
	prefix.Write(hashPrefix)
	res := make(hash.Events, 0, 10)

	it := s.table.Events.NewIterator(prefix.Bytes(), nil)
	defer it.Release()
	for it.Next() {
		res = append(res, hash.BytesToEvent(it.Key()))
	}

	return res
}

// GetEventPayloadRLP returns stored event. Serialized.
func (s *Store) GetEventPayloadRLP(id hash.Event) rlp.RawValue {
	key := id.Bytes()

	data, err := s.table.Events.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	return data
}

// HasEvent returns true if event exists.
func (s *Store) HasEvent(h hash.Event) bool {
	if has, ok := s.cache.EventIDs.Has(h); ok {
		return has
	}
	has, _ := s.table.Events.Has(h.Bytes())
	return has
}

func (s *Store) loadHighestLamport() idx.Lamport {
	lamportBytes, err := s.table.HighestLamport.Get([]byte("k"))
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if lamportBytes == nil {
		return 0
	}
	return idx.BytesToLamport(lamportBytes)
}

func (s *Store) getCachedHighestLamport() (idx.Lamport, bool) {
	cache := s.cache.HighestLamport.Load()
	if cache != nil {
		return cache.(idx.Lamport), true
	}
	return 0, false
}

func (s *Store) GetHighestLamport() idx.Lamport {
	cached, ok := s.getCachedHighestLamport()
	if ok {
		return cached
	}
	lamport := s.loadHighestLamport()
	s.cache.HighestLamport.Store(lamport)
	return lamport
}

func (s *Store) SetHighestLamport(lamport idx.Lamport) {
	s.cache.HighestLamport.Store(lamport)
}

func (s *Store) FlushHighestLamport() {
	cached, ok := s.getCachedHighestLamport()
	if !ok {
		return
	}
	err := s.table.HighestLamport.Put([]byte("k"), cached.Bytes())
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}
