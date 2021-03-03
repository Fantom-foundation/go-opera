package evmstore

import "github.com/syndtr/goleveldb/leveldb/opt"

type (
	// StoreCacheConfig is a config for the db.
	StoreCacheConfig struct {
		// Cache size for Receipts (size in bytes).
		ReceiptsSize uint
		// Cache size for Receipts (number of blocks).
		ReceiptsBlocks int
		// Cache size for TxPositions.
		TxPositions int
		EvmDatabase int
	}
	// StoreConfig is a config for store db.
	StoreConfig struct {
		Cache StoreCacheConfig
	}
)

// DefaultStoreConfig for product.
func DefaultStoreConfig() StoreConfig {
	return StoreConfig{
		StoreCacheConfig{
			ReceiptsSize:   64 * 1024,
			ReceiptsBlocks: 1000,
			TxPositions:    5000,
			EvmDatabase:    16 * opt.MiB,
		},
	}
}

// LiteStoreConfig is for tests or inmemory.
func LiteStoreConfig() StoreConfig {
	return StoreConfig{
		StoreCacheConfig{
			ReceiptsSize:   3 * 1024,
			ReceiptsBlocks: 100,
			TxPositions:    500,
		},
	}
}
