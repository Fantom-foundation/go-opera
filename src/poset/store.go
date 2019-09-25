package poset

import (
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

// Store is a poset persistent storage working over physical key-value database.
type Store struct {
	persistentDB kvdb.KeyValueStore
	table        struct {
		Checkpoint     kvdb.KeyValueStore `table:"checkpoint_"`
		Epochs         kvdb.KeyValueStore `table:"epoch_"`
		ConfirmedEvent kvdb.KeyValueStore `table:"confirmed_"`
		FrameInfos     kvdb.KeyValueStore `table:"frameinfo_"`
	}

	epochDb    kvdb.KeyValueStore
	epochTable struct {
		Roots       kvdb.KeyValueStore `table:"roots_"`
		VectorIndex kvdb.KeyValueStore `table:"vectors_"`
	}

	makeDb func(name string) kvdb.KeyValueStore

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.KeyValueStore, makeDb func(name string) kvdb.KeyValueStore) *Store {
	s := &Store{
		persistentDB: db,
		epochDb:      makeDb("epoch"),
		makeDb:       makeDb,
		Instance:     logger.MakeInstance(),
	}

	table.MigrateTables(&s.table, s.persistentDB)
	table.MigrateTables(&s.epochTable, s.epochDb)

	return s
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	return NewStore(memorydb.New(), func(name string) kvdb.KeyValueStore {
		return memorydb.New()
	})
}

// Close leaves underlying database.
func (s *Store) Close() {
	table.MigrateTables(&s.table, nil)
	table.MigrateTables(&s.epochTable, nil)
	err := s.persistentDB.Close()
	if err != nil {
		s.Log.Crit("Failed to close persistent db", "err", err)
	}
	err = s.epochDb.Close()
	if err != nil {
		s.Log.Crit("Failed to close epoch db", "err", err)
	}
}

func (s *Store) recreateEpochDb() {
	if s.epochDb != nil {
		err := s.epochDb.Close()
		if err != nil {
			s.Log.Crit("Failed to close epoch db", "err", err)
		}
		s.epochDb.Drop()
	}
	s.epochDb = s.makeDb("epoch")
	table.MigrateTables(&s.epochTable, s.epochDb)
}

/*
 * Utils:
 */

func (s *Store) set(table kvdb.KeyValueStore, key []byte, val interface{}) {
	buf, err := rlp.EncodeToBytes(val)
	if err != nil {
		s.Log.Crit("Failed to encode rlp", "err", err)
	}

	if err := table.Put(key, buf); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (s *Store) get(table kvdb.KeyValueStore, key []byte, to interface{}) interface{} {
	buf, err := table.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return nil
	}

	err = rlp.DecodeBytes(buf, to)
	if err != nil {
		s.Log.Crit("Failed to decode rlp", "err", err, "size", len(buf))
	}
	return to
}

func (s *Store) has(table kvdb.KeyValueStore, key []byte) bool {
	res, err := table.Has(key)
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	return res
}
