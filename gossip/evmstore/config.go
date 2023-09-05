package evmstore

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type (
	// StoreCacheConfig is a config for the db.
	StoreCacheConfig struct {
		// Cache size for Receipts (size in bytes).
		ReceiptsSize uint
		// Cache size for Receipts (number of blocks).
		ReceiptsBlocks int
		// Cache size for TxPositions.
		TxPositions int
		// Cache size for EVM database.
		EvmDatabase int
		// Cache size for EVM snapshot.
		EvmSnap int
		// Cache size for EvmBlock (number of blocks).
		EvmBlocksNum int
		// Cache size for EvmBlock (size in bytes).
		EvmBlocksSize uint
		// Disk journal for saving clean cache entries.
		TrieCleanJournal string
		// Whether to disable trie write caching and GC altogether (archive node)
		TrieDirtyDisabled bool
		// Memory limit (MB) at which to start flushing dirty trie nodes to disk
		TrieDirtyLimit int
		// Memory allowance (MB) to use for caching trie nodes in memory
		TrieCleanLimit int
		// Time limit after which to flush the current in-memory trie to disk
		TrieTimeLimit time.Duration
		// Memory allowance (MB) to use for caching snapshot entries in memory
		SnapshotLimit int
		// Whether to enable greedy gc mode
		GreedyGC bool
		// Number of blocks from head whose state histories are reserved.
		StateHistory uint64
		// Scheme used to store ethereum states and merkle tree nodes on top
		StateScheme string
	}
	// StoreConfig is a config for store db.
	StoreConfig struct {
		Cache StoreCacheConfig
		// Enables tracking of SHA3 preimages in the VM
		EnablePreimageRecording bool
	}
)

// DefaultStoreConfig for product.
func DefaultStoreConfig(scale cachescale.Func) StoreConfig {
	return StoreConfig{
		Cache: StoreCacheConfig{
			ReceiptsSize:      scale.U(4 * opt.MiB),
			ReceiptsBlocks:    scale.I(4000),
			TxPositions:       scale.I(20000),
			EvmDatabase:       scale.I(32 * opt.MiB),
			EvmSnap:           scale.I(32 * opt.MiB),
			EvmBlocksNum:      scale.I(5000),
			EvmBlocksSize:     scale.U(6 * opt.MiB),
			TrieDirtyDisabled: true,
			GreedyGC:          false,
			TrieDirtyLimit:    scale.I(256 * opt.MiB),
			TrieCleanLimit:    scale.I(128 * opt.MiB),
			TrieTimeLimit:     5 * time.Minute,
			StateScheme:       rawdb.HashScheme,
			StateHistory:      params.FullImmutabilityThreshold,
		},
		EnablePreimageRecording: true,
	}
}

// LiteStoreConfig is for tests or inmemory.
func LiteStoreConfig() StoreConfig {
	return DefaultStoreConfig(cachescale.Ratio{Base: 10, Target: 1})
}
