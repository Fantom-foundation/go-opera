package posposet

import (
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// Store is a poset persistent storage working over physical key-value database.
type Store struct {
	physicalDB kvdb.Database

	// TODO: cache with LRU.
	states kvdb.Database
	events kvdb.Database
	frames kvdb.Database
	blocks kvdb.Database

	balances state.Database
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	s := &Store{
		physicalDB: kvdb.NewMemDatabase(),
	}
	s.init()
	return s
}

// NewBadgerStore creates store over badger database.
func NewBadgerStore(db *badger.DB) *Store {
	s := &Store{
		physicalDB: kvdb.NewBadgerDatabase(db),
	}
	s.init()
	return s
}

func (s *Store) init() {
	s.states = kvdb.NewTable(s.physicalDB, "state_")
	s.events = kvdb.NewTable(s.physicalDB, "event_")
	s.frames = kvdb.NewTable(s.physicalDB, "frame_")
	s.blocks = kvdb.NewTable(s.physicalDB, "block_")

	s.balances = state.NewDatabase(
		kvdb.NewTable(s.physicalDB, "balance_"))
}

// Close leaves underlying database.
func (s *Store) Close() {
	s.states = nil
	s.events = nil
	s.frames = nil
	s.balances = nil
	s.physicalDB.Close()
}

// ApplyGenesis stores initial state.
func (s *Store) ApplyGenesis(balances map[hash.Address]uint64) error {
	st := s.GetState()
	if st != nil {
		return fmt.Errorf("Genesis has applied already")
	}

	if balances == nil {
		return fmt.Errorf("Balances shouldn't be nil")
	}

	st = &State{
		LastFinishedFrameN: 0,
		TotalCap:           0,
	}

	genesis := s.StateDB(hash.Hash{})
	for addr, balance := range balances {
		genesis.AddBalance(hash.Address(addr), balance)
		st.TotalCap += balance
	}

	if st.TotalCap < uint64(len(balances)) {
		return fmt.Errorf("Balance shouldn't be zero")
	}

	var err error
	st.Genesis, err = genesis.Commit(true)
	if err != nil {
		return err
	}

	s.SetState(st)
	return nil
}

// SetEvent stores event.
func (s *Store) SetEvent(e *Event) {
	s.set(s.events, e.Hash().Bytes(), e.ToWire())
}

// GetEvent returns stored event.
func (s *Store) GetEvent(h hash.EventHash) *Event {
	w, _ := s.get(s.events, h.Bytes(), &wire.Event{}).(*wire.Event)
	e := WireToEvent(w)
	if e != nil {
		e.hash = h // fill cache
	}
	return e
}

// HasEvent returns true if event exists.
func (s *Store) HasEvent(h hash.EventHash) bool {
	return s.has(s.events, h.Bytes())
}

// SetState stores state.
func (s *Store) SetState(st *State) {
	const key = "current"
	s.set(s.states, []byte(key), st.ToWire())
}

// GetState returns stored state.
func (s *Store) GetState() *State {
	const key = "current"
	w, _ := s.get(s.states, []byte(key), &wire.State{}).(*wire.State)
	return WireToState(w)
}

// SetFrame stores event.
func (s *Store) SetFrame(f *Frame) {
	s.set(s.frames, intToKey(f.Index), f.ToWire())
}

// GetFrame returns stored frame.
func (s *Store) GetFrame(n uint64) *Frame {
	w, _ := s.get(s.frames, intToKey(n), &wire.Frame{}).(*wire.Frame)
	return WireToFrame(w)
}

// SetBlock stores chain block.
func (s *Store) SetBlock(b *Block) {
	s.set(s.blocks, intToKey(b.Index), b.ToWire())
}

// GetBlock returns stored block.
func (s *Store) GetBlock(n uint64) *Block {
	w, _ := s.get(s.blocks, intToKey(n), &wire.Block{}).(*wire.Block)
	return WireToBlock(w)
}

// StateDB returns state database.
func (s *Store) StateDB(from hash.Hash) *state.DB {
	db, err := state.New(from, s.balances)
	if err != nil {
		panic(err)
	}
	return db
}

/*
 * Utils:
 */

func (s *Store) set(table kvdb.Database, key []byte, val proto.Message) {
	var pbf proto.Buffer
	pbf.SetDeterministic(true)

	if err := pbf.Marshal(val); err != nil {
		panic(err)
	}

	if err := table.Put(key, pbf.Bytes()); err != nil {
		panic(err)
	}
}

func (s *Store) get(table kvdb.Database, key []byte, to proto.Message) proto.Message {
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

func (s *Store) has(table kvdb.Database, key []byte) bool {
	res, err := table.Has(key)
	if err != nil {
		panic(err)
	}
	return res
}

func intToKey(n uint64) []byte {
	var res [8]byte
	for i := 0; i < len(res); i++ {
		res[i] = byte(n)
		n = n >> 8
	}
	return res[:]
}
