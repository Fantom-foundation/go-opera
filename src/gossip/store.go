package gossip

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/no_key_is_err"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	persistentDB kvdb.KeyValueStore

	table struct {
		Peers     kvdb.KeyValueStore `table:"peer_"`
		Events    kvdb.KeyValueStore `table:"event_"`
		Blocks    kvdb.KeyValueStore `table:"block_"`
		PackInfos kvdb.KeyValueStore `table:"packinfo_"`
		Packs     kvdb.KeyValueStore `table:"pack_"`
		PacksNum  kvdb.KeyValueStore `table:"packs_num_"`

		TmpDbs kvdb.KeyValueStore `table:"tmpdbs_"`

		Evm      ethdb.Database
		EvmState state.Database
	}

	tmpDbs

	makeDb func(name string) kvdb.KeyValueStore

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.KeyValueStore, makeDb func(name string) kvdb.KeyValueStore) *Store {
	s := &Store{
		persistentDB: db,
		makeDb:       makeDb,
		Instance:     logger.MakeInstance(),
	}

	table.MigrateTables(&s.table, s.persistentDB)

	evmTable := no_key_is_err.Wrap(table.New(s.persistentDB, []byte("evm_"))) // ETH expects that "not found" is an error
	s.table.Evm = rawdb.NewDatabase(evmTable)
	s.table.EvmState = state.NewDatabase(s.table.Evm)

	s.initTmpDbs()

	return s
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	db := memorydb.New()
	return NewStore(db, func(name string) kvdb.KeyValueStore {
		return memorydb.New()
	})
}

// Close leaves underlying database.
func (s *Store) Close() {
	table.MigrateTables(&s.table, nil)
	s.persistentDB.Close()
}

// StateDB returns state database.
func (s *Store) StateDB(from common.Hash) *state.StateDB {
	db, err := state.New(common.Hash(from), s.table.EvmState)
	if err != nil {
		s.Log.Crit("Failed to open state", "err", err)
	}
	return db
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
