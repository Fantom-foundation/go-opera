package poset

// StoreConfig is a config for store db.
type StoreConfig struct {
	// Cache size for Roots.
	Roots int
}

// DefaultStoreConfig for product.
func DefaultStoreConfig() StoreConfig {
	return StoreConfig{
		Roots: 20,
	}
}

// LiteStoreConfig is for tests or inmemory.
func LiteStoreConfig() StoreConfig {
	return StoreConfig{
		Roots: 5,
	}
}
