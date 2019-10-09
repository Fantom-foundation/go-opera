package poset

import (
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

const (
	immediately = true
)

// Store is a poset persistent storage working over parent key-value database.
type Store struct {
	dbs *flushable.SyncedPool

	mainDb kvdb.KeyValueStore
	table  struct {
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

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(dbs *flushable.SyncedPool) *Store {
	s := &Store{
		dbs:      dbs,
		mainDb:   dbs.GetDb("poset-main"),
		Instance: logger.MakeInstance(),
	}

	table.MigrateTables(&s.table, s.mainDb)

	return s
}

// NewMemStore creates store over memory map.
// Store is always blank.
func NewMemStore() *Store {
	mems := memorydb.NewProdicer("")
	dbs := flushable.NewSyncedPool(mems)

	return NewStore(dbs)
}

// Commit changes hard.
func (s *Store) Commit(e hash.Event, immediately bool) {
	if immediately {
		s.dbs.Flush(e.Bytes())
	} else {
		s.dbs.FlushIfNeeded(e.Bytes())
	}
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
		prevDb = s.dbs.GetDb(name(n - 1))
	}

	err := prevDb.Close()
	if err != nil {
		s.Log.Crit("Failed to close epoch db", "err", err)
	}
	prevDb.Drop()

	s.epochDb = s.dbs.GetDb(name(n))
	table.MigrateTables(&s.epochTable, s.epochDb)
}

func name(n idx.Epoch) string {
	return fmt.Sprintf("poset-epoch-%d", n)
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
