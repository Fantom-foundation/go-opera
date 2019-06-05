package posposet

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

const cacheSize = 500 // TODO: Move it to config later

// Store is a poset persistent storage working over physical key-value database.
type Store struct {
	physicalDB kvdb.Database

	table struct {
		States      kvdb.Database `table:"state_"`
		Frames      kvdb.Database `table:"frame_"`
		Blocks      kvdb.Database `table:"block_"`
		Event2Frame kvdb.Database `table:"event2frame_"`
		Event2Block kvdb.Database `table:"event2block_"`
		Balances    state.Database
	}
	cache struct {
		Frames      *lru.Cache `cache:"-"`
		Event2Frame *lru.Cache `cache:"-"`
		Event2Block *lru.Cache `cache:"-"`
	}

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.Database, cached bool) *Store {
	s := &Store{
		physicalDB: db,
		Instance:   logger.MakeInstance(),
	}

	kvdb.MigrateTables(&s.table, s.physicalDB)
	s.table.Balances = state.NewDatabase(kvdb.NewTable(s.physicalDB, "balance_"))

	if cached {
		kvdb.MigrateCaches(&s.cache, func() interface{} {
			c, err := lru.New(cacheSize)
			if err != nil {
				s.Fatal(err)
			}
			return c
		})
	}

	return s
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	db := kvdb.NewMemDatabase()
	return NewStore(db, false)
}

// Close leaves underlying database.
func (s *Store) Close() {
	kvdb.MigrateCaches(&s.cache, nil)
	kvdb.MigrateTables(&s.table, nil)
	s.physicalDB.Close()
}

// ApplyGenesis stores initial state.
func (s *Store) ApplyGenesis(balances map[hash.Peer]uint64) error {
	if balances == nil {
		return fmt.Errorf("balances shouldn't be nil")
	}

	st := s.GetState()
	if st != nil {
		if st.Genesis == genesisHash(balances) {
			return nil
		}
		return fmt.Errorf("other genesis has applied already")
	}

	st = &State{
		LastFinishedFrameN: 0,
		TotalCap:           0,
	}

	genesis := s.StateDB(hash.Hash{})
	for addr, balance := range balances {
		genesis.SetBalance(hash.Peer(addr), balance)
		st.TotalCap += balance
	}

	if st.TotalCap < uint64(len(balances)) {
		return fmt.Errorf("balance shouldn't be zero")
	}

	var err error
	st.Genesis, err = genesis.Commit(true)
	if err != nil {
		return err
	}

	s.SetState(st)
	return nil
}

/*
 * Utils:
 */

func (s *Store) set(table kvdb.Database, key []byte, val proto.Message) {
	var pbf proto.Buffer

	if err := pbf.Marshal(val); err != nil {
		s.Fatal(err)
	}

	if err := table.Put(key, pbf.Bytes()); err != nil {
		s.Fatal(err)
	}
}

func (s *Store) get(table kvdb.Database, key []byte, to proto.Message) proto.Message {
	buf, err := table.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return nil
	}

	err = proto.Unmarshal(buf, to)
	if err != nil {
		s.Fatal(err)
	}
	return to
}

func (s *Store) has(table kvdb.Database, key []byte) bool {
	res, err := table.Has(key)
	if err != nil {
		s.Fatal(err)
	}
	return res
}
