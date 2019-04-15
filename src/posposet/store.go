package posposet

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// Store is a poset persistent storage working over physical key-value database.
// TODO: cache tables with LRU.
type Store struct {
	physicalDB kvdb.Database

	// indexes:
	event2frame kvdb.Database

	// entities:
	states kvdb.Database
	frames kvdb.Database
	blocks kvdb.Database

	// trie:
	balances state.Database
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.Database) *Store {
	s := &Store{
		physicalDB: db,
	}
	s.init()
	return s
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	db := kvdb.NewMemDatabase()
	return NewStore(db)
}

func (s *Store) init() {
	s.event2frame = kvdb.NewTable(s.physicalDB, "event2frame_")

	s.states = kvdb.NewTable(s.physicalDB, "state_")
	s.frames = kvdb.NewTable(s.physicalDB, "frame_")
	s.blocks = kvdb.NewTable(s.physicalDB, "block_")

	s.balances = state.NewDatabase(
		kvdb.NewTable(s.physicalDB, "balance_"))
}

// Close leaves underlying database.
func (s *Store) Close() {
	s.balances = nil
	s.frames = nil
	s.states = nil
	s.event2frame = nil
	s.physicalDB.Close()
}

// ApplyGenesis stores initial state.
func (s *Store) ApplyGenesis(balances map[hash.Peer]uint64) error {
	if balances == nil {
		return fmt.Errorf("Balances shouldn't be nil")
	}

	st := s.GetState()
	if st != nil {
		if st.Genesis == GenesisHash(balances) {
			return nil
		}
		return fmt.Errorf("Other genesis has applied already")
	}

	st = &State{
		LastFinishedFrameN: 0,
		TotalCap:           0,
	}

	genesis := s.StateDB(hash.Hash{})
	for addr, balance := range balances {
		genesis.AddBalance(hash.Peer(addr), balance)
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

// SetState stores frame num of event.
func (s *Store) SetEventFrame(e hash.Event, frame uint64) {
	key := e.Bytes()
	val := intToBytes(frame)
	if err := s.event2frame.Put(key, val); err != nil {
		panic(err)
	}
}

// GetEventFrame returns frame num of event.
func (s *Store) GetEventFrame(e hash.Event) *uint64 {
	key := e.Bytes()
	buf, err := s.event2frame.Get(key)
	if err != nil {
		panic(err)
	}
	if buf == nil {
		return nil
	}

	val := bytesToInt(buf)
	return &val
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
	s.set(s.frames, intToBytes(f.Index), f.ToWire())
}

// GetFrame returns stored frame.
func (s *Store) GetFrame(n uint64) *Frame {
	w, _ := s.get(s.frames, intToBytes(n), &wire.Frame{}).(*wire.Frame)
	return WireToFrame(w)
}

// SetBlock stores chain block.
func (s *Store) SetBlock(b *Block) {
	s.set(s.blocks, intToBytes(b.Index), b.ToWire())
}

// GetBlock returns stored block.
func (s *Store) GetBlock(n uint64) *Block {
	w, _ := s.get(s.blocks, intToBytes(n), &wire.Block{}).(*wire.Block)
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

func intToBytes(n uint64) []byte {
	var res [8]byte
	for i := 0; i < len(res); i++ {
		res[i] = byte(n)
		n = n >> 8
	}
	return res[:]
}

func bytesToInt(b []byte) uint64 {
	var res uint64
	for i := 0; i < len(b); i++ {
		res += uint64(b[i]) << uint(i*8)
	}
	return res
}
