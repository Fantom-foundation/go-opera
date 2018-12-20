package node

import (
	"testing"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/log"
	"github.com/sirupsen/logrus"
)

// Config for node configuration settings
type Config struct {
	HeartbeatTimeout time.Duration `mapstructure:"heartbeat"`
	TCPTimeout       time.Duration `mapstructure:"timeout"`
	CacheSize        int           `mapstructure:"cache-size"`
	SyncLimit        int64         `mapstructure:"sync-limit"`
	Logger           *logrus.Logger
	TestDelay uint64 `mapstructure:"test_delay"`
}

// NewConfig creates a new node config
func NewConfig(heartbeat time.Duration,
	timeout time.Duration,
	cacheSize int,
	syncLimit int64,
	logger *logrus.Logger) *Config {

	return &Config{
		HeartbeatTimeout: heartbeat,
		TCPTimeout:       timeout,
		CacheSize:        cacheSize,
		SyncLimit:        syncLimit,
		Logger:           logger,
	}
}

// DefaultConfig sets the default config for a node config
func DefaultConfig() *Config {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	lachesis_log.NewLocal(logger, logger.Level.String())

	return &Config{
		HeartbeatTimeout: 10 * time.Millisecond,
		TCPTimeout:       180 * 1000 * time.Millisecond,
		CacheSize:        500,
		SyncLimit:        100,
		Logger:           logger,
		TestDelay:        1,
	}
}

// TestConfig sets the test config for use with tests
func TestConfig(t *testing.T) *Config {
	config := DefaultConfig()
	config.HeartbeatTimeout = time.Second * 1

	config.Logger = common.NewTestLogger(t)

	return config
}
