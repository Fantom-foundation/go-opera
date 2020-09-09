package app

type (
	// StoreConfig is a config for store db.
	StoreConfig struct {
		// Cache size for Receipts.
		ReceiptsCacheSize int
		// Cache size for Stakers.
		StakersCacheSize int
	}
)

// DefaultStoreConfig for product.
func DefaultStoreConfig() StoreConfig {
	return StoreConfig{
		ReceiptsCacheSize: 100,
		StakersCacheSize:  4000,
	}
}

// LiteStoreConfig is for tests or inmemory.
func LiteStoreConfig() StoreConfig {
	return StoreConfig{
		ReceiptsCacheSize: 100,
		StakersCacheSize:  400,
	}
}
