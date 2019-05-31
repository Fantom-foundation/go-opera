package mobile

// Config stores all the configuration information for a mobile node
type Config struct {
	Heartbeat  int    //heartbeat timeout in milliseconds
	TCPTimeout int    //TCP timeout in milliseconds
	MaxPool    int    //Max number of pooled connections
	CacheSize  int    //Number of items in LRU cache
	SyncLimit  int    //Max Events per sync
	StoreType  string //inmem or badger
	StorePath  string //File containing the Store DB
}

// NewMobileConfig creates a new mobile config
func NewMobileConfig(heartbeat int,
	tcpTimeout int,
	maxPool int,
	cacheSize int,
	syncLimit int,
	storeType string,
	storePath string) *Config {

	return &Config{
		Heartbeat:  heartbeat,
		TCPTimeout: tcpTimeout,
		MaxPool:    maxPool,
		CacheSize:  cacheSize,
		SyncLimit:  syncLimit,
		StoreType:  storeType,
		StorePath:  storePath,
	}
}

// DefaultConfig sets the default config
func DefaultConfig() *Config {
	return &Config{
		Heartbeat:  1000,
		TCPTimeout: 1000,
		MaxPool:    2,
		CacheSize:  500,
		SyncLimit:  1000,
		StoreType:  "inmem",
		StorePath:  "",
	}
}
