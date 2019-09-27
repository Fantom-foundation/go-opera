package poset

import (
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

// Store is a poset persistent storage working over parent key-value database.
type Store struct {
	mainDb *flushable.Flushable
	table  struct {
		Checkpoint     kvdb.KeyValueStore `table:"checkpoint_"`
		Event2Block    kvdb.KeyValueStore `table:"event2block_"`
		Epochs         kvdb.KeyValueStore `table:"epoch_"`
		ConfirmedEvent kvdb.KeyValueStore `table:"confirmed_"`
		FrameInfos     kvdb.KeyValueStore `table:"frameinfo_"`
	}

	epochDb    *flushable.Flushable
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
		mainDb:   flushable.New(db),
		makeDb:   makeDb,
		Instance: logger.MakeInstance(),
	}

	table.MigrateTables(&s.table, s.mainDb)

	return s
}

// NewMemStore creates store over memory map.
// Store is always blank.
func NewMemStore() *Store {
	return NewStore(memorydb.New(), func(name string) kvdb.KeyValueStore {
		return memorydb.New()
	})
}

// Commit changes.
func (s *Store) Commit() error {
	err := s.epochDb.Flush()
	if err != nil {
		return err
	}

	return s.mainDb.Flush()
}

// Close leaves underlying database.
func (s *Store) Close() {
	table.MigrateTables(&s.table, nil)
	table.MigrateTables(&s.epochTable, nil)
	err := s.mainDb.Close()
	if err != nil {
		s.Log.Crit("Failed to close persistent db", "err", err)
	}
	err = s.epochDb.Close()
	if err != nil {
		s.Log.Crit("Failed to close epoch db", "err", err)
	}
}

// RecreateEpochDb makes new epoch DB and drops prev.
func (s *Store) RecreateEpochDb(n idx.Epoch) {
	prevDb := s.epochDb
	if prevDb == nil {
		prevDb = flushable.New(s.makeDb(name(n - 1)))
	}
	err := prevDb.Close()
	if err != nil {
		s.Log.Crit("Failed to close epoch db", "err", err)
	}
	prevDb.Drop()

	s.epochDb = flushable.New(s.makeDb(name(n)))
	table.MigrateTables(&s.epochTable, s.epochDb)
}

func name(n idx.Epoch) string {
	return fmt.Sprintf("epoch-%d", n)
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
		s.Log.Crit("Failed to decode rlp", "err", err)
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
