package posposet

import (
	"github.com/dgraph-io/badger"

	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/rlp"
)

// Store is a poset persistent storage working over physical key-value database.
// TODO: make it internal.
type Store struct {
	physicalDB kvdb.Database

	states kvdb.Database
	events kvdb.Database
}

// NewInmemStore creates store over memory map.
func NewMemStore() *Store {
	s := &Store{
		physicalDB: kvdb.NewMemDatabase(),
	}
	s.init()
	return s
}

// NewInmemStore creates store over badger database.
func NewBadgerStore(db *badger.DB) *Store {
	s := &Store{
		physicalDB: kvdb.NewBadgerDatabase(db),
	}
	s.init()
	return s
}

func (s *Store) init() {
	s.states = kvdb.NewTable(s.physicalDB, "states_")
	s.events = kvdb.NewTable(s.physicalDB, "events_")
}

// Close leaves underlying database.
func (s *Store) Close() {
	s.states = nil
	s.events = nil
	s.physicalDB.Close()
}

// SetEvent stores event.
func (s *Store) SetEvent(e *Event) {
	s.set(s.events, e.Hash().Bytes(), e)
}

// GetEvent returns stored event.
func (s *Store) GetEvent(h EventHash) *Event {
	e, _ := s.get(s.events, h.Bytes(), &Event{}).(*Event)
	return e
}

// HasEvent returns true if event exists.
func (s *Store) HasEvent(h EventHash) bool {
	return s.has(s.events, h.Bytes())
}

// SetEvent stores event.
func (s *Store) SetState(st *State) {
	const key = "current"
	s.set(s.states, []byte(key), st)
}

// GetEvent returns stored event.
func (s *Store) GetState() *State {
	const key = "current"
	st, _ := s.get(s.states, []byte(key), &State{}).(*State)
	return st
}

/*
 * Utils:
 */

func (s *Store) set(table kvdb.Database, key []byte, val interface{}) {
	buf, err := rlp.EncodeToBytes(val)
	if err != nil {
		panic(err)
	}

	err = table.Put(key, buf)
	if err != nil {
		panic(err)
	}
}

func (s *Store) get(table kvdb.Database, key []byte, to interface{}) interface{} {
	buf, err := table.Get(key)
	if err != nil {
		panic(err)
	}
	if buf == nil {
		return nil
	}

	err = rlp.DecodeBytes(buf, to)
	if err != nil {
		panic(err)
	}
	return to
}

func (s *Store) has(table kvdb.Database, key []byte) bool {
	res, err := table.Has(key)
	if err != nil {
		panic(err)
	}
	return res
}
