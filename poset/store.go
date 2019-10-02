package poset

import (
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

// Store is a poset persistent storage working over parent key-value database.
type Store struct {
	bareDb kvdb.KeyValueStore

	mainDb *flushable.Flushable
	table  struct {
		Checkpoint     kvdb.KeyValueStore `table:"checkpoint_"`
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
		bareDb:   db,
		mainDb:   flushable.New(db),
		makeDb:   makeDb,
		Instance: logger.MakeInstance(),
	}

	table.MigrateTables(&s.table, s.mainDb)

	if s.isDirty() {
		s.Log.Crit("Consensus DB is possible inconsistent. Recreate it.")
	}

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
	s.setDirty(true)
	defer s.setDirty(false)

	err := s.epochDb.Flush()
	if err != nil {
		return err
	}

	return s.mainDb.Flush()
}

// setDirty sets dirty flag.
func (s *Store) setDirty(flag bool) {
	key := []byte("is_dirty")
	val := make([]byte, 1, 1)
	if flag {
		val[0] = 1
	}

	err := s.bareDb.Put(key, val)
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// isDirty gets dirty flag.
func (s *Store) isDirty() bool {
	key := []byte("is_dirty")
	val, err := s.bareDb.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get value", "err", err)
	}

	return len(val) > 1 && val[0] != 0
}

// Close leaves underlying database.
func (s *Store) Close() {
	table.MigrateTables(&s.table, nil)
	table.MigrateTables(&s.epochTable, nil)
	err := s.mainDb.Close()
	if err != nil {
		s.Log.Crit("Failed to close persistent db", "err", err)
	}

	if s.epochDb == nil {
		return
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
