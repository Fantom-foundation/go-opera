package posposet

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
)

// EventStore is a poset event storage for test purpose.
// It implements EventSource interface.
type EventStore struct {
	physicalDB kvdb.Database
	events     kvdb.Database
}

// NewEventStore creates store over memory map.
func NewEventStore() *EventStore {
	s := &EventStore{
		physicalDB: kvdb.NewMemDatabase(),
	}
	s.init()
	return s
}

func (s *EventStore) init() {
	s.events = kvdb.NewTable(s.physicalDB, "event_")
}

// Close leaves underlying database.
func (s *EventStore) Close() {
	s.events = nil
	s.physicalDB.Close()
}

// SetEvent stores event.
func (s *EventStore) SetEvent(e *inter.Event) {
	s.set(s.events, e.Hash().Bytes(), e.ToWire())
}

// GetEvent returns stored event.
func (s *EventStore) GetEvent(h hash.Event) *inter.Event {
	w, _ := s.get(s.events, h.Bytes(), &wire.Event{}).(*wire.Event)
	e := inter.WireToEvent(w)
	return e
}

// HasEvent returns true if event exists.
func (s *EventStore) HasEvent(h hash.Event) bool {
	return s.has(s.events, h.Bytes())
}

/*
 * Tests:
 */

func TestEventStore(t *testing.T) {
	store := NewEventStore()

	t.Run("NotExisting", func(t *testing.T) {
		assert := assert.New(t)

		h := hash.FakeEvent()
		e1 := store.GetEvent(h)
		assert.Nil(e1)
	})

	t.Run("Events", func(t *testing.T) {
		assert := assert.New(t)

		events := inter.FakeFuzzingEvents()
		for _, e0 := range events {
			store.SetEvent(e0)
			e1 := store.GetEvent(e0.Hash())

			if !assert.Equal(e0.Hash(), e1.Hash()) {
				break
			}
			if !assert.Equal(e0, e1) {
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
