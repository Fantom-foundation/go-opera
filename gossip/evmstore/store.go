package evmstore

import (
	"errors"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/nokeyiserr"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/Fantom-foundation/lachesis-base/utils/wlru"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/prque"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/topicsdb"
	"github.com/Fantom-foundation/go-opera/utils/adapters/kvdb2ethdb"
	"github.com/Fantom-foundation/go-opera/utils/rlpstore"
)

const nominalSize uint = 1

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	cfg StoreConfig

	table struct {
		Evm kvdb.Store `table:"M"`
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

	rlp rlpstore.Helper

	triegc *prque.Prque // Priority queue mapping block numbers to tries to gc

	logger.Instance
}

const (
	TriesInMemory = 16
)

// NewStore creates store over key-value db.
func NewStore(dbs kvdb.DBProducer, cfg StoreConfig) *Store {
	s := &Store{
		cfg:      cfg,
		Instance: logger.New("evm-store"),
		rlp:      rlpstore.Helper{logger.New("rlp")},
		triegc:   prque.New(nil),
	}

	err := table.OpenTables(&s.table, dbs, "evm")
	if err != nil {
		s.Log.Crit("Failed to open tables", "err", err)
	}

	s.initEVMDB()
	s.EvmLogs = topicsdb.New(dbs)
	s.initCache()

	return s
}

// Close closes underlying database.
func (s *Store) Close() {
	setnil := func() interface{} {
		return nil
	}

	_ = table.CloseTables(&s.table)
	table.MigrateTables(&s.table, nil)
	table.MigrateCaches(&s.cache, setnil)
	s.EvmLogs.Close()
}

func (s *Store) initCache() {
	s.cache.Receipts = s.makeCache(s.cfg.Cache.ReceiptsSize, s.cfg.Cache.ReceiptsBlocks)
	s.cache.TxPositions = s.makeCache(nominalSize*uint(s.cfg.Cache.TxPositions), s.cfg.Cache.TxPositions)
	s.cache.EvmBlocks = s.makeCache(s.cfg.Cache.EvmBlocksSize, s.cfg.Cache.EvmBlocksNum)
}

func (s *Store) initEVMDB() {
	s.EvmDb = rawdb.NewDatabase(
		kvdb2ethdb.Wrap(
			nokeyiserr.Wrap(
				s.table.Evm)))
	s.EvmState = state.NewDatabaseWithConfig(s.EvmDb, &trie.Config{
		Cache:     s.cfg.Cache.EvmDatabase / opt.MiB,
		Journal:   s.cfg.Cache.TrieCleanJournal,
		Preimages: s.cfg.EnablePreimageRecording,
		GreedyGC:  s.cfg.Cache.GreedyGC,
	})
}

func (s *Store) ResetWithEVMDB(evmStore kvdb.Store) *Store {
	cp := *s
	cp.table.Evm = evmStore
	cp.initEVMDB()
	cp.Snaps = nil
	return &cp
}

func (s *Store) EVMDB() kvdb.Store {
	return s.table.Evm
}

func (s *Store) GenerateEvmSnapshot(root common.Hash, rebuild, async bool) (err error) {
	if s.Snaps != nil {
		return errors.New("EVM snapshot is already opened")
	}
	s.Snaps, err = snapshot.New(
		s.EvmDb,
		s.EvmState.TrieDB(),
		s.cfg.Cache.EvmSnap/opt.MiB,
		root,
		async,
		rebuild,
		false)
	return
}

func (s *Store) RebuildEvmSnapshot(root common.Hash) {
	if s.Snaps == nil {
		return
	}
	s.Snaps.Rebuild(root)
}

// CleanCommit clean old state trie and commit changes.
func (s *Store) CleanCommit(block iblockproc.BlockState) error {
	// Don't need to reference the current state root
	// due to it already be referenced on `Commit()` function
	triedb := s.EvmState.TrieDB()
	stateRoot := common.Hash(block.FinalizedStateRoot)
	if current := uint64(block.LastBlock.Idx); current > TriesInMemory {
		// Find the next state trie we need to commit
		chosen := current - TriesInMemory
		// Garbage collect all below the chosen block
		for !s.triegc.Empty() {
			root, number := s.triegc.Pop()
			if uint64(-number) > chosen {
				s.triegc.Push(root, number)
				break
			}
			triedb.Dereference(root.(common.Hash))
		}
	}
	// commit the state trie after clean up
	err := triedb.Commit(stateRoot, false, nil)
	if err != nil {
		s.Log.Error("Failed to flush trie DB into main DB", "err", err)
	}
	return err
}

func (s *Store) PauseEvmSnapshot() {
	s.Snaps.Disable()
}

func (s *Store) IsEvmSnapshotPaused() bool {
	return rawdb.ReadSnapshotDisabled(s.table.Evm)
}

// Commit changes.
func (s *Store) Commit(block idx.Block, root hash.Hash, flush bool) error {
	triedb := s.EvmState.TrieDB()
	stateRoot := common.Hash(root)
	// If we're applying genesis or running an archive node, always flush
	if flush || s.cfg.Cache.TrieDirtyDisabled {
		err := triedb.Commit(stateRoot, false, nil)
		if err != nil {
			s.Log.Error("Failed to flush trie DB into main DB", "err", err)
		}
		return err
	} else {
		// Full but not archive node, do proper garbage collection
		triedb.Reference(stateRoot, common.Hash{}) // metadata reference to keep trie alive
		s.triegc.Push(stateRoot, -int64(block))

		if current := uint64(block); current > TriesInMemory {
			// If we exceeded our memory allowance, flush matured singleton nodes to disk
			s.Cap()

			// Find the next state trie we need to commit
			chosen := current - TriesInMemory

			// Garbage collect all below the chosen block
			for !s.triegc.Empty() {
				root, number := s.triegc.Pop()
				if uint64(-number) > chosen {
					s.triegc.Push(root, number)
					break
				}
				triedb.Dereference(root.(common.Hash))
			}
		}
		return nil
	}
}

func (s *Store) Flush(block iblockproc.BlockState) {
	// Ensure that the entirety of the state snapshot is journalled to disk.
	var snapBase common.Hash
	if s.Snaps != nil {
		var err error
		if snapBase, err = s.Snaps.Journal(common.Hash(block.FinalizedStateRoot)); err != nil {
			s.Log.Error("Failed to journal state snapshot", "err", err)
		}
	}
	// Ensure the state of a recent block is also stored to disk before exiting.
	if !s.cfg.Cache.TrieDirtyDisabled {
		triedb := s.EvmState.TrieDB()

		if number := uint64(block.LastBlock.Idx); number > 0 {
			s.Log.Info("Writing cached state to disk", "block", number, "root", block.FinalizedStateRoot)
			if err := triedb.Commit(common.Hash(block.FinalizedStateRoot), true, nil); err != nil {
				s.Log.Error("Failed to commit recent state trie", "err", err)
			}
		}
		if snapBase != (common.Hash{}) {
			s.Log.Info("Writing snapshot state to disk", "root", snapBase)
			if err := triedb.Commit(snapBase, true, nil); err != nil {
				s.Log.Error("Failed to commit recent state trie", "err", err)
			}
		}
	}
	// Ensure all live cached entries be saved into disk, so that we can skip
	// cache warmup when node restarts.
	if s.cfg.Cache.TrieCleanJournal != "" {
		triedb := s.EvmState.TrieDB()
		triedb.SaveCache(s.cfg.Cache.TrieCleanJournal)
	}
}

// Cap flush matured singleton nodes to disk
func (s *Store) Cap() {
	triedb := s.EvmState.TrieDB()
	var (
		nodes, imgs = triedb.Size()
		limit       = common.StorageSize(s.cfg.Cache.TrieDirtyLimit)
	)
	// If we exceeded our memory allowance, flush matured singleton nodes to disk
	if nodes > limit+ethdb.IdealBatchSize || imgs > 4*1024*1024 {
		triedb.Cap(limit)
	}
}

// StateDB returns state database.
func (s *Store) StateDB(from hash.Hash) (*state.StateDB, error) {
	return state.NewWithSnapLayers(common.Hash(from), s.EvmState, s.Snaps, 0)
}

// HasStateDB returns if state database exists
func (s *Store) HasStateDB(from hash.Hash) bool {
	_, err := s.StateDB(from)
	return err == nil
}

// IndexLogs indexes EVM logs
func (s *Store) IndexLogs(recs ...*types.Log) {
	err := s.EvmLogs.Push(recs...)
	if err != nil {
		s.Log.Crit("DB logs index error", "err", err)
	}
}

func (s *Store) Snapshots() *snapshot.Tree {
	return s.Snaps
}

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
