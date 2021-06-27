package evmstore

import (
	"sync"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/nokeyiserr"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/Fantom-foundation/lachesis-base/utils/wlru"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/topicsdb"
	"github.com/Fantom-foundation/go-opera/utils/adapters/kvdb2ethdb"
	"github.com/Fantom-foundation/go-opera/utils/rlpstore"
)

const nominalSize uint = 1

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	cfg StoreConfig

	mainDB kvdb.Store
	table  struct {
		Evm  kvdb.Store `table:"M"`
		Logs kvdb.Store `table:"L"`
		// API-only tables
		Receipts    kvdb.Store `table:"r"`
		TxPositions kvdb.Store `table:"x"`
		Txs         kvdb.Store `table:"X"`
	}

	EvmDb    ethdb.Database
	EvmState state.Database
	EvmLogs  *topicsdb.Index
	Snaps    *snapshot.Tree

	cache struct {
		TxPositions *wlru.Cache `cache:"-"` // store by pointer
		Receipts    *wlru.Cache `cache:"-"` // store by value
		EvmBlocks   *wlru.Cache `cache:"-"` // store by pointer
	}

	mutex struct {
		Inc sync.Mutex
	}

	rlp rlpstore.Helper

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(mainDB kvdb.Store, cfg StoreConfig) *Store {
	s := &Store{
		cfg:      cfg,
		mainDB:   mainDB,
		Instance: logger.New("evm-store"),
		rlp:      rlpstore.Helper{logger.New("rlp")},
	}

	table.MigrateTables(&s.table, s.mainDB)

	s.EvmDb = rawdb.NewDatabase(
		kvdb2ethdb.Wrap(
			nokeyiserr.Wrap(
				s.table.Evm)))
	s.EvmState = state.NewDatabaseWithConfig(s.EvmDb, &trie.Config{
		Cache:     cfg.Cache.EvmDatabase / opt.MiB,
		Preimages: cfg.EnablePreimageRecording,
	})
	s.EvmLogs = topicsdb.New(s.table.Logs)

	s.initCache()

	return s
}

func (s *Store) initCache() {
	s.cache.Receipts = s.makeCache(s.cfg.Cache.ReceiptsSize, s.cfg.Cache.ReceiptsBlocks)
	s.cache.TxPositions = s.makeCache(nominalSize*uint(s.cfg.Cache.TxPositions), s.cfg.Cache.TxPositions)
	s.cache.EvmBlocks = s.makeCache(s.cfg.Cache.EvmBlocksSize, s.cfg.Cache.EvmBlocksNum)
}

func (s *Store) InitEvmSnapshot(root hash.Hash) (err error) {
	s.Snaps, err = snapshot.New(
		kvdb2ethdb.Wrap(nokeyiserr.Wrap(s.table.Evm)),
		s.EvmState.TrieDB(),
		s.cfg.Cache.EvmSnap/opt.MiB,
		common.Hash(root),
		false, true, false)
	return err
}

// Commit changes.
func (s *Store) Commit(root hash.Hash) error {
	// Flush trie on the DB
	err := s.EvmState.TrieDB().Commit(common.Hash(root), false, nil)
	if err != nil {
		s.Log.Error("Failed to flush trie DB into main DB", "err", err)
	}
	return err
}

func (s *Store) Cap(max, min int) {
	maxSize := common.StorageSize(max)
	minSize := common.StorageSize(min)
	size, preimagesSize := s.EvmState.TrieDB().Size()
	if size >= maxSize || preimagesSize >= maxSize {
		_ = s.EvmState.TrieDB().Cap(minSize)
	}
}

// StateDB returns state database.
func (s *Store) StateDB(from hash.Hash) (*state.StateDB, error) {
	return state.NewWithSnapLayers(common.Hash(from), s.EvmState, s.Snaps, 0)
}

// IndexLogs indexes EVM logs
func (s *Store) IndexLogs(recs ...*types.Log) {
	err := s.EvmLogs.Push(recs...)
	if err != nil {
		s.Log.Crit("DB logs index error", "err", err)
	}
}

/*
func (s *Store) EvmTable() ethdb.Database {
	return s.table.EvmDb
}

func (s *Store) EvmDatabase() state.Database {
	return s.table.EvmState
}

func (s *Store) EvmLogs() *topicsdb.Index {
	return s.table.EvmLogs
}
*/

/*
 * Utils:
 */

func (s *Store) makeCache(weight uint, size int) *wlru.Cache {
	cache, err := wlru.New(weight, size)
	if err != nil {
		s.Log.Crit("Failed to create LRU cache", "err", err)
		return nil
	}
	return cache
}
