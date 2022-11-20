package evmstore

import (
	"errors"
	"sync"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/Fantom-foundation/lachesis-base/utils/wlru"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/prque"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/topicsdb"

	"github.com/Fantom-foundation/go-opera/utils/rlpstore"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore/ethdb"
	erigonethdb "github.com/ledgerwatch/erigon/ethdb"

	"github.com/Fantom-foundation/go-opera/utils/adapters/kvdb2ethdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/nokeyiserr"
	"github.com/ethereum/go-ethereum/core/rawdb"
	gethethdb "github.com/ethereum/go-ethereum/ethdb"
	"github.com/ledgerwatch/erigon-lib/kv"
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

	chainKV  kv.RwDB
	EvmDb    erigonethdb.Database
	EvmState state.Database // includes caching mechanism (DBStateReader)
	EvmLogs  *topicsdb.Index

	LegacyEvmDb gethethdb.Database

	cache struct {
		TxPositions *wlru.Cache `cache:"-"` // store by pointer
		Receipts    *wlru.Cache `cache:"-"` // store by value
		EvmBlocks   *wlru.Cache `cache:"-"` // store by pointer
	}

	mutex struct {
		Inc sync.Mutex
	}

	rlp rlpstore.Helper

	triegc *prque.Prque // Priority queue mapping block numbers to tries to gc

	logger.Instance
}

const (
	TriesInMemory = 32
)

// NewStore creates store over key-value db.
func NewStore(mainDB kvdb.Store, cfg StoreConfig, chainKV kv.RwDB) *Store {
	s := &Store{
		cfg:      cfg,
		mainDB:   mainDB,
		Instance: logger.New("evm-store"),
		rlp:      rlpstore.Helper{logger.New("rlp")},
		triegc:   prque.New(nil),
	}

	table.MigrateTables(&s.table, s.mainDB)
	s.chainKV = chainKV
	s.initEVMDB(chainKV) // consider to add genesisKV as well, or merge genesisKV and chainKV
	s.EvmLogs = topicsdb.New(s.table.Logs)

	s.LegacyEvmDb = rawdb.NewDatabase(
		kvdb2ethdb.Wrap(
			nokeyiserr.Wrap(
				s.table.Evm)))

	s.initCache()

	return s
}

func (s *Store) initEVMDB(chainKV kv.RwDB) {
	s.EvmDb = ethdb.NewObjectDatabase(chainKV)
	s.EvmState = state.NewDatabase(s.EvmDb)
}

func (s *Store) initCache() {
	s.cache.Receipts = s.makeCache(s.cfg.Cache.ReceiptsSize, s.cfg.Cache.ReceiptsBlocks)
	s.cache.TxPositions = s.makeCache(nominalSize*uint(s.cfg.Cache.TxPositions), s.cfg.Cache.TxPositions)
	s.cache.EvmBlocks = s.makeCache(s.cfg.Cache.EvmBlocksSize, s.cfg.Cache.EvmBlocksNum)
}

func (s *Store) Close() {
	//s.genesisKV.Close()
	//s.chainKV.Close()
}

func (s *Store) GenerateEvmSnapshot(root common.Hash, rebuild, async bool) (err error) {
	return errors.New("EVM snapshot is not implemented yet")
}

func (s *Store) RebuildEvmSnapshot(root common.Hash) {
	/*
		if s.Snaps == nil {
			return
		}
		s.Snaps.Rebuild(root)
	*/
	// TODO: deal with snaps
}

func (s *Store) PauseEvmSnapshot() {
	//s.Snaps.Disable()
	// TODO: deal with snaps
}

func (s *Store) IsEvmSnapshotPaused() bool {
	// return rawdb.ReadSnapshotDisabled(s.table.Evm)
	// TODO: deal with snaps
	return false
}

func (s *Store) SnapsGenerating() (bool, error) {
	// TODO: deal with snaps
	return false, nil
}

// Commit changes.

/*
func (s *Store) Commit(block iblockproc.BlockState, flush bool) error {
	triedb := s.EvmState.TrieDB()
	stateRoot := common.Hash(block.FinalizedStateRoot)
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
		s.triegc.Push(stateRoot, -int64(block.LastBlock.Idx))

		if current := uint64(block.LastBlock.Idx); current > TriesInMemory {
			// If we exceeded our memory allowance, flush matured singleton nodes to disk
			var (
				nodes, imgs = triedb.Size()
				limit       = common.StorageSize(s.cfg.Cache.TrieDirtyLimit)
			)
			if nodes > limit || imgs > 4*1024*1024 {
				triedb.Cap(limit - ethdb.IdealBatchSize)
			}
			// Find the next state trie we need to commit
			chosen := current - TriesInMemory

			// Garbage collect anything below our required write retention
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
*/

/*
func (s *Store) Flush(block iblockproc.BlockState, getBlock func(n idx.Block) *inter.Block) {
	// Ensure that the entirety of the state snapshot is journalled to disk.
	var snapBase common.Hash
	if s.Snaps != nil {
		var err error
		if snapBase, err = s.Snaps.Journal(common.Hash(block.FinalizedStateRoot)); err != nil {
			s.Log.Error("Failed to journal state snapshot", "err", err)
		}
	}
	// Ensure the state of a recent block is also stored to disk before exiting.
	// We're writing three different states to catch different restart scenarios:
	//  - HEAD:     So we don't need to reprocess any blocks in the general case
	//  - HEAD-1:   So we don't do large reorgs if our HEAD becomes an uncle
	//  - HEAD-31:  So we have a hard limit on the number of blocks reexecuted
	if !s.cfg.Cache.TrieDirtyDisabled {
		triedb := s.EvmState.TrieDB()

		for _, offset := range []uint64{0, 1, TriesInMemory - 1} {
			if number := uint64(block.LastBlock.Idx); number > offset {
				recent := getBlock(idx.Block(number - offset))
				if recent == nil || recent.Root == hash.Zero {
					break
				}
				s.Log.Info("Writing cached state to disk", "block", number-offset, "root", recent.Root)
				if err := triedb.Commit(common.Hash(recent.Root), true, nil); err != nil {
					s.Log.Error("Failed to commit recent state trie", "err", err)
				}
			}
		}
		if snapBase != (common.Hash{}) {
			s.Log.Info("Writing snapshot state to disk", "root", snapBase)
			if err := triedb.Commit(snapBase, true, nil); err != nil {
				s.Log.Error("Failed to commit recent state trie", "err", err)
			}
		}
		for !s.triegc.Empty() {
			triedb.Dereference(s.triegc.PopItem().(common.Hash))
		}
		if size, _ := triedb.Size(); size != 0 {
			s.Log.Error("Dangling trie nodes after full cleanup")
		}
	}
	// Ensure all live cached entries be saved into disk, so that we can skip
	// cache warmup when node restarts.
	if s.cfg.Cache.TrieCleanJournal != "" {
		triedb := s.EvmState.TrieDB()
		triedb.SaveCache(s.cfg.Cache.TrieCleanJournal)
	}
}
*/

/*
func (s *Store) Cap(max, min int) {
	maxSize := common.StorageSize(max)
	minSize := common.StorageSize(min)
	size, preimagesSize := s.EvmState.TrieDB().Size()
	if size >= maxSize || preimagesSize >= maxSize {
		_ = s.EvmState.TrieDB().Cap(minSize)
	}
}
*/

// StateDB returns state database.
func (s *Store) StateDB(from hash.Hash) *state.StateDB {
	return state.NewWithRoot(common.Hash(from))
}

// HasStateDB returns if state database exists
/*
func (s *Store) HasStateDB(from hash.Hash) bool {
	_, err := s.StateDB(from)
	return err == nil
}
*/

// IndexLogs indexes EVM logs
func (s *Store) IndexLogs(recs ...*types.Log) {
	err := s.EvmLogs.Push(recs...)
	if err != nil {
		s.Log.Crit("DB logs index error", "err", err)
	}
}

func (s *Store) Snapshots() *snapshot.Tree {
	// TODO: deal with snaps
	return nil
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

func (s *Store) ChainKV() kv.RwDB {
	return s.chainKV
}
