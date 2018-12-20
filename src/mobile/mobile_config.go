package mobile

// MobileConfig stores all the configuration information for a mobile node
type MobileConfig struct {
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
	storePath string) *MobileConfig {

	return &MobileConfig{
		Heartbeat:  heartbeat,
		TCPTimeout: tcpTimeout,
		MaxPool:    maxPool,
		CacheSize:  cacheSize,
		SyncLimit:  syncLimit,
		StoreType:  storeType,
		StorePath:  storePath,
	}
}

// DefaultMobileConfig sets the default config
func DefaultMobileConfig() *MobileConfig {
	return &MobileConfig{
		Heartbeat:  1000,
		TCPTimeout: 1000,
		MaxPool:    2,
		CacheSize:  500,
		SyncLimit:  1000,
		StoreType:  "inmem",
		StorePath:  "",
	}
}
