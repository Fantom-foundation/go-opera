package posposet

import (
	"github.com/dgraph-io/badger"

	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/rlp"
)

// Store is a poset persistent storage.
type Store struct {
	physicalDB kvdb.Database

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
	s.events = kvdb.NewTable(s.physicalDB, "events_")
}

// Close leaves underlying database.
func (s *Store) Close() {
	s.events = nil
	s.physicalDB.Close()
}

func (s *Store) SetEvent(e *Event) {
	hash := e.Hash()
	s.set(hash.Bytes(), e)
}

func (s *Store) GetEvent(h EventHash) *Event {
	val, _ := s.get(h.Bytes(), &Event{}).(*Event)
	return val
}

/*
 * Utils:
 */

func (s *Store) set(key []byte, val interface{}) {
	buf, err := rlp.EncodeToBytes(val)
	if err != nil {
		panic(err)
	}

	err = s.events.Put(key, buf)
	if err != nil {
		panic(err)
	}
}

func (s *Store) get(key []byte, to interface{}) interface{} {
	buf, err := s.events.Get(key)
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
