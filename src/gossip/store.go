package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/state"

	"github.com/ethereum/go-ethereum/rlp"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	physicalDB kvdb.Database

	table struct {
		Peers     kvdb.Database `table:"peer_"`
		Events    kvdb.Database `table:"event_"`
		Blocks    kvdb.Database `table:"block_"`
		PackInfos kvdb.Database `table:"packinfo_"`
		Packs     kvdb.Database `table:"pack_"`
		PacksNum  kvdb.Database `table:"packs_num_"`
		Balances  state.Database
		Headers   kvdb.Database `table:"header_"` // TODO should be temporary, epoch-scoped
		Tips      kvdb.Database `table:"tips_"`   // TODO should be temporary, epoch-scoped
		Heads     kvdb.Database `table:"heads_"`  // TODO should be temporary, epoch-scoped
	}

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.Database) *Store {
	s := &Store{
		physicalDB: db,
		Instance:   logger.MakeInstance(),
	}

	kvdb.MigrateTables(&s.table, s.physicalDB)
	s.table.Balances = state.NewDatabase(
		s.physicalDB.NewTable([]byte("balance_")))

	return s
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	db := kvdb.NewMemDatabase()
	return NewStore(db)
}

// Close leaves underlying database.
func (s *Store) Close() {
	kvdb.MigrateTables(&s.table, nil)
	s.physicalDB.Close()
}

// StateDB returns state database.
func (s *Store) StateDB(from hash.Hash) *state.DB {
	db, err := state.New(from, s.table.Balances)
	if err != nil {
		s.Fatal(err)
	}
	return db
}

/*
 * Utils:
 */

func (s *Store) set(table kvdb.Database, key []byte, val interface{}) {
	buf, err := rlp.EncodeToBytes(val)
	if err != nil {
		s.Fatal(err)
	}

	if err := table.Put(key, buf); err != nil {
		s.Fatal(err)
	}
}

func (s *Store) get(table kvdb.Database, key []byte, to interface{}) interface{} {
	buf, err := table.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return nil
	}

	err = rlp.DecodeBytes(buf, to)
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
