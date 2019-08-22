package gossip

import (
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	persistentDB kvdb.Database

	table struct {
		Peers     kvdb.Database `table:"peer_"`
		Events    kvdb.Database `table:"event_"`
		Blocks    kvdb.Database `table:"block_"`
		PackInfos kvdb.Database `table:"packinfo_"`
		Packs     kvdb.Database `table:"pack_"`
		PacksNum  kvdb.Database `table:"packs_num_"`

		TmpDbs kvdb.Database `table:"tmpdbs_"`

		Balances state.Database
	}

	tmpDbs

	makeDb func(name string) kvdb.Database

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.Database, makeDb func(name string) kvdb.Database) *Store {
	s := &Store{
		persistentDB: db,
		makeDb:       makeDb,
		Instance:     logger.MakeInstance(),
	}

	kvdb.MigrateTables(&s.table, s.persistentDB)
	s.table.Balances = state.NewDatabase(
		s.persistentDB.NewTable([]byte("balance_")))

	s.initTmpDbs()

	return s
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	db := kvdb.NewMemDatabase()
	return NewStore(db, func(name string) kvdb.Database {
		return kvdb.NewMemDatabase()
	})
}

// Close leaves underlying database.
func (s *Store) Close() {
	kvdb.MigrateTables(&s.table, nil)
	s.persistentDB.Close()
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
