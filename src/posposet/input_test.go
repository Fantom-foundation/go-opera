package posposet

import (
	"testing"

	"github.com/golang/protobuf/proto"
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
	physicalDB kvdb.Database

	table struct {
		Events  kvdb.Database `table:"event_"`
		ExtTxns kvdb.Database `table:"exttxn_"`
	}

	logger.Instance
}

// NewEventStore creates store over memory map.
func NewEventStore(db kvdb.Database) *EventStore {
	if db == nil {
		db = kvdb.NewMemDatabase()
	}

	s := &EventStore{
		physicalDB: db,
		Instance:   logger.MakeInstance(),
	}

	kvdb.MigrateTables(&s.table, s.physicalDB)

	return s
}

// Close leaves underlying database.
func (s *EventStore) Close() {
	kvdb.MigrateTables(&s.table, nil)
	s.physicalDB.Close()
}

// SetEvent stores event.
func (s *EventStore) SetEvent(e *inter.Event) {
	w, wt := e.ToWire()
	s.set(s.table.ExtTxns, e.Hash().Bytes(), wt.ExtTxnsValue)
	s.set(s.table.Events, e.Hash().Bytes(), w)
}

// GetEvent returns stored event.
func (s *EventStore) GetEvent(h hash.Event) *inter.Event {
	w, _ := s.get(s.table.Events, h.Bytes(), &wire.Event{}).(*wire.Event)
	if w == nil {
		return nil
	}
	wt, _ := s.get(s.table.ExtTxns, h.Bytes(), &wire.ExtTxns{}).(*wire.ExtTxns)
	w.ExternalTransactions = &wire.Event_ExtTxnsValue{ExtTxnsValue: wt}

	return inter.WireToEvent(w)
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

	store := NewEventStore(nil)

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
