package poset

import (
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/ethereum/go-ethereum/rlp"
)

// Store is a poset persistent storage working over physical key-value database.
type Store struct {
	persistentDB kvdb.Database
	table        struct {
		Checkpoint     kvdb.Database `table:"checkpoint_"`
		Event2Block    kvdb.Database `table:"event2block_"`
		SuperFrames    kvdb.Database `table:"sframe_"`
		ConfirmedEvent kvdb.Database `table:"confirmed_"`
		FrameInfos     kvdb.Database `table:"frameinfo_"`
	}

	epochDb    kvdb.Database
	epochTable struct {
		Roots       kvdb.Database `table:"roots_"`
		VectorIndex kvdb.Database `table:"vectors_"`
	}

	makeDb func(name string) kvdb.Database

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.Database, makeDb func(name string) kvdb.Database) *Store {
	s := &Store{
		persistentDB: db,
		epochDb:      makeDb("epoch"),
		makeDb:       makeDb,
		Instance:     logger.MakeInstance(),
	}

	kvdb.MigrateTables(&s.table, s.persistentDB)
	kvdb.MigrateTables(&s.epochTable, s.epochDb)

	return s
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	return NewStore(kvdb.NewMemDatabase(), func(name string) kvdb.Database {
		return kvdb.NewMemDatabase()
	})
}

// Close leaves underlying database.
func (s *Store) Close() {
	kvdb.MigrateTables(&s.table, nil)
	kvdb.MigrateTables(&s.epochTable, nil)
	s.persistentDB.Close()
	s.epochDb.Close()
}

func (s *Store) recreateEpochDb() {
	if s.epochDb != nil {
		s.epochDb.Close()
		s.epochDb.Drop()
	}
	s.epochDb = s.makeDb("epoch")
	kvdb.MigrateTables(&s.epochTable, s.epochDb)
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
