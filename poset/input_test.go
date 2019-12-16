package poset

import (
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	lru "github.com/hashicorp/golang-lru"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

// EventStore is a poset event storage for test purpose.
// It implements EventSource interface.
type EventStore struct {
	physicalDB kvdb.KeyValueStore

	table struct {
		Events kvdb.KeyValueStore `table:"e"`
	}
	cache struct {
		Events *lru.Cache `cache:"-"` // store by pointer
	}

	logger.Instance
}

// NewEventStore creates store over memory map.
func NewEventStore(db kvdb.KeyValueStore, cacheSize int) *EventStore {
	if db == nil {
		db = memorydb.New()
	}

	s := &EventStore{
		physicalDB: db,
		Instance:   logger.MakeInstance(),
	}

	s.cache.Events = s.makeCache(cacheSize)

	table.MigrateTables(&s.table, s.physicalDB)

	return s
}

// Close leaves underlying database.
func (s *EventStore) Close() {
	table.MigrateTables(&s.table, nil)
	s.physicalDB.Close()
}

// SetEvent stores event.
func (s *EventStore) SetEvent(e *inter.Event) {
	s.set(s.table.Events, e.Hash().Bytes(), e)
	s.cache.Events.Add(e.Hash(), e)
}

// GetEvent returns stored event.
func (s *EventStore) GetEvent(h hash.Event) *inter.Event {
	if val, ok := s.cache.Events.Get(h); ok {
		return val.(*inter.Event)
	}

	w, _ := s.get(s.table.Events, h.Bytes(), &inter.Event{}).(*inter.Event)
	if w == nil {
		return nil
	}

	return w
}

// GetEventHeader returns stored event header.
// Note: fake epoch partition.
func (s *EventStore) GetEventHeader(_ idx.Epoch, h hash.Event) *inter.EventHeaderData {
	e := s.GetEvent(h)
	if e == nil {
		return nil
	}
	return &e.EventHeaderData
}

// HasEvent returns true if event exists.
func (s *EventStore) HasEvent(h hash.Event) bool {
	return s.has(s.table.Events, h.Bytes())
}

/*
 * Tests:
 */

func TestEventStore(t *testing.T) {
	logger.SetTestMode(t)

	store := NewEventStore(nil, 100)

	t.Run("NotExisting", func(t *testing.T) {
		assertar := assert.New(t)

		h := hash.FakeEvent()
		e1 := store.GetEvent(h)
		assertar.Nil(e1)
	})

	t.Run("Events", func(t *testing.T) {
		assertar := assert.New(t)

		events := inter.FakeFuzzingEvents()
		for _, e0 := range events {
			store.SetEvent(e0)
			e1 := store.GetEvent(e0.Hash())

			if !assertar.Equal(e0.Hash(), e1.Hash()) {
				break
			}
			if !assertar.Equal(e0, e1) {
				break
			}
		}
	})

	store.Close()
}

/*
 * Utils:
 */

func (s *EventStore) set(table kvdb.KeyValueStore, key []byte, val interface{}) {
	buf, err := rlp.EncodeToBytes(val)
	if err != nil {
		s.Log.Crit("Failed to encode rlp", "err", err)
	}

	if err := table.Put(key, buf); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (s *EventStore) get(table kvdb.KeyValueStore, key []byte, to interface{}) interface{} {
	buf, err := table.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return nil
	}

	err = rlp.DecodeBytes(buf, to)
	if err != nil {
		s.Log.Crit("Failed to decode rlp", "err", err)
	}
	return to
}

func (s *EventStore) has(table kvdb.KeyValueStore, key []byte) bool {
	res, err := table.Has(key)
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	return res
}

func (s *EventStore) makeCache(size int) *lru.Cache {
	if size <= 0 {
		return nil
	}

	cache, err := lru.New(size)
	if err != nil {
		s.Log.Crit("Error create LRU cache", "err", err)
		return nil
	}
	return cache
}
