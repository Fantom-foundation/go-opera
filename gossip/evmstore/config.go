package evmstore

type (
	// StoreConfig is a config for store db.
	StoreConfig struct {
		// Cache size for Receipts.
		ReceiptsCacheSize int
		// Cache size for TxPositions.
		TxPositionsCacheSize int
	}
)

// DefaultStoreConfig for product.
func DefaultStoreConfig() StoreConfig {
	return StoreConfig{
		ReceiptsCacheSize:    5,
		TxPositionsCacheSize: 1000,
	}
}

// LiteStoreConfig is for tests or inmemory.
func LiteStoreConfig() StoreConfig {
	return StoreConfig{
		ReceiptsCacheSize:    1,
		TxPositionsCacheSize: 100,
	}
}
