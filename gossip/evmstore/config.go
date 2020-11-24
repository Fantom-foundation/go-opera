package evmstore

type (
	// StoreConfig is a config for store db.
	StoreCacheConfig struct {
		// Cache size for Receipts (size in bytes).
		ReceiptsSize uint
		// Cache size for Receipts (number of blocks).
		ReceiptsBlocks int
		// Cache size for TxPositions.
		TxPositions int
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
