package posposet

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/golang-lru"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

// EventStore is a poset event storage for test purpose.
// It implements EventSource interface.
type EventStore struct {
	physicalDB  kvdb.Database
	events      kvdb.Database
	eventsCache *lru.Cache
}

// NewEventStore creates store over memory map.
func NewEventStore(db kvdb.Database, cached bool) *EventStore {
	if db == nil {
		db = kvdb.NewMemDatabase()
	}

	s := &EventStore{
		physicalDB: db,
		events:     kvdb.NewTable(db, "event_"),
	}

	if cached {
		c, err := lru.New(cacheSize)
		if err != nil {
			panic(err)
		}
		s.eventsCache = c
	}

	return s
}

// Close leaves underlying database.
func (s *EventStore) Close() {
	if s.eventsCache != nil {
		s.eventsCache.Purge()
	}
	s.events = nil
	s.physicalDB.Close()
}

// SetEvent stores event.
func (s *EventStore) SetEvent(e *inter.Event) {
	w := e.ToWire()
	s.set(s.events, e.Hash().Bytes(), w)

	if s.eventsCache != nil {
		s.eventsCache.Add(e.Hash(), w)
	}
}

// GetEvent returns stored event.
func (s *EventStore) GetEvent(h hash.Event) *inter.Event {
	if s.eventsCache != nil {
		if e, ok := s.eventsCache.Get(h); ok {
			w := e.(*wire.Event)
			return inter.WireToEvent(w)
		}
	}

	w, _ := s.get(s.events, h.Bytes(), &wire.Event{}).(*wire.Event)
	return inter.WireToEvent(w)
}

// HasEvent returns true if event exists.
func (s *EventStore) HasEvent(h hash.Event) bool {
	if s.eventsCache.Contains(h) {
		return true
	}

	return s.has(s.events, h.Bytes())
}

/*
 * Tests:
 */

func TestEventStore(t *testing.T) {
	logger.SetTestMode(t)

	store := NewEventStore(nil, false)

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

func (s *EventStore) set(table kvdb.Database, key []byte, val proto.Message) {
	var pbf proto.Buffer
	pbf.SetDeterministic(true)

	if err := pbf.Marshal(val); err != nil {
		panic(err)
	}

	if err := table.Put(key, pbf.Bytes()); err != nil {
		panic(err)
	}
}

func (s *EventStore) get(table kvdb.Database, key []byte, to proto.Message) proto.Message {
	buf, err := table.Get(key)
	if err != nil {
		panic(err)
	}
	if buf == nil {
		return nil
	}

	err = proto.Unmarshal(buf, to)
	if err != nil {
		panic(err)
	}
	return to
}

func (s *EventStore) has(table kvdb.Database, key []byte) bool {
	res, err := table.Has(key)
	if err != nil {
		panic(err)
	}
	return res
}
