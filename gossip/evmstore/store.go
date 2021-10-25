package evmstore

import (
	"sync"
	"sync/atomic"

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

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/inter"
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
		// API-only tables
		Receipts    kvdb.Store `table:"r"`
		TxPositions kvdb.Store `table:"x"`
		Txs         kvdb.Store `table:"X"`

		Evm      ethdb.Database
		EvmState state.Database
		EvmLogs  *topicsdb.Index
		Snaps    *snapshot.Tree
	}

	cache struct {
		TxPositions *wlru.Cache `cache:"-"` // store by pointer
		Receipts    *wlru.Cache `cache:"-"` // store by value
		EvmBlocks   *wlru.Cache `cache:"-"` // store by pointer
	}

	mutex struct {
		Inc sync.Mutex
	}

	rlp rlpstore.Helper

	currentBlock atomic.Value // Current head of the block chain

	snaps  *snapshot.Tree // Snapshot tree for fast trie leaf access
	triegc *prque.Prque   // Priority queue mapping block numbers to tries to gc

	logger.Instance
}

const (
	TriesInMemory = 32
)

// NewStore creates store over key-value db.
func NewStore(mainDB kvdb.Store, cfg StoreConfig) *Store {
	s := &Store{
		cfg:      cfg,
		mainDB:   mainDB,
		Instance: logger.New("evm-store"),
		rlp:      rlpstore.Helper{logger.New("rlp")},
		triegc:   prque.New(nil),
	}

	table.MigrateTables(&s.table, s.mainDB)

	evmTable := nokeyiserr.Wrap(s.EvmKvdbTable()) // ETH expects that "not found" is an error
	s.table.Evm = rawdb.NewDatabase(kvdb2ethdb.Wrap(evmTable))
	s.table.EvmState = state.NewDatabaseWithConfig(s.table.Evm, &trie.Config{
		Cache:     cfg.Cache.EvmDatabase / opt.MiB,
		Journal:   cfg.Cache.TrieCleanJournal,
		Preimages: cfg.EnablePreimageRecording,
	})
	s.table.EvmLogs = topicsdb.New(table.New(s.mainDB, []byte("L")))

	s.initCache()

	var nilBlock *evmcore.EvmBlock
	s.currentBlock.Store(nilBlock)

	return s
}

func (s *Store) initCache() {
	s.cache.Receipts = s.makeCache(s.cfg.Cache.ReceiptsSize, s.cfg.Cache.ReceiptsBlocks)
	s.cache.TxPositions = s.makeCache(nominalSize*uint(s.cfg.Cache.TxPositions), s.cfg.Cache.TxPositions)
	s.cache.EvmBlocks = s.makeCache(s.cfg.Cache.EvmBlocksSize, s.cfg.Cache.EvmBlocksNum)
}

func (s *Store) InitEvmSnapshot(root hash.Hash) (err error) {
	s.table.Snaps, err = snapshot.New(kvdb2ethdb.Wrap(nokeyiserr.Wrap(s.EvmKvdbTable())), s.table.EvmState.TrieDB(), s.cfg.Cache.EvmSnap/opt.MiB, common.Hash(root), false, true, false)
	return err
}

// Commit changes.
func (s *Store) Commit(block *evmcore.EvmBlock, genesis bool) error {
	triedb := s.table.EvmState.TrieDB()
	// If we're applying genesis or running an archive node, always flush
	if genesis == true || s.cfg.Cache.TrieDirtyDisabled {
		err := triedb.Commit(block.Root, false, nil)
		if err != nil {
			s.Log.Error("Failed to flush trie DB into main DB", "err", err)
		}
		s.currentBlock.Store(block)
		return err
	} else {
		// Full but not archive node, do proper garbage collection
		triedb.Reference(block.Root, common.Hash{}) // metadata reference to keep trie alive
		s.triegc.Push(block.Root, -int64(block.NumberU64()))

		if current := block.NumberU64(); current > TriesInMemory {
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
		s.currentBlock.Store(block)
		return nil
	}
}

func (s *Store) Flush(getBlock func(n idx.Block) *inter.Block) {
	// Ensure that the entirety of the state snapshot is journalled to disk.
	var snapBase common.Hash
	if s.snaps != nil {
		var err error
		if snapBase, err = s.snaps.Journal(s.CurrentBlock().Root); err != nil {
			s.Log.Error("Failed to journal state snapshot", "err", err)
		}
	}
	// Ensure the state of a recent block is also stored to disk before exiting.
	// We're writing three different states to catch different restart scenarios:
	//  - HEAD:     So we don't need to reprocess any blocks in the general case
	//  - HEAD-1:   So we don't do large reorgs if our HEAD becomes an uncle
	//  - HEAD-127: So we have a hard limit on the number of blocks reexecuted
	if !s.cfg.Cache.TrieDirtyDisabled {
		triedb := s.table.EvmState.TrieDB()

		for _, offset := range []uint64{0, 1, TriesInMemory - 1} {
			if number := s.CurrentBlock().NumberU64(); number > offset {
				recent := getBlock(idx.Block(number - offset))
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
		triedb := s.table.EvmState.TrieDB()
		triedb.SaveCache(s.cfg.Cache.TrieCleanJournal)
	}
}

// CurrentBlock retrieves the current head block of the canonical chain. The
// block is retrieved from the blockchain's internal cache.
func (s *Store) CurrentBlock() *evmcore.EvmBlock {
	return s.currentBlock.Load().(*evmcore.EvmBlock)
}

func (s *Store) Cap(max, min int) {
	maxSize := common.StorageSize(max)
	minSize := common.StorageSize(min)
	size, preimagesSize := s.table.EvmState.TrieDB().Size()
	if size >= maxSize || preimagesSize >= maxSize {
		_ = s.table.EvmState.TrieDB().Cap(minSize)
	}
}

// StateDB returns state database.
func (s *Store) StateDB(from hash.Hash) (*state.StateDB, error) {
	return state.NewWithSnapLayers(common.Hash(from), s.table.EvmState, s.table.Snaps, 0)
}

// IndexLogs indexes EVM logs
func (s *Store) IndexLogs(recs ...*types.Log) {
	err := s.table.EvmLogs.Push(recs...)
	if err != nil {
		s.Log.Crit("DB logs index error", "err", err)
	}
}

func (s *Store) EvmKvdbTable() kvdb.Store {
	return table.New(s.mainDB, []byte("M"))
}

func (s *Store) EvmTable() ethdb.Database {
	return s.table.Evm
}

func (s *Store) EvmDatabase() state.Database {
	return s.table.EvmState
}

func (s *Store) EvmLogs() *topicsdb.Index {
	return s.table.EvmLogs
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
