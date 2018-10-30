package node

import (
	"testing"
	"time"

	"github.com/andrecronje/lachesis/src/common"
	"github.com/andrecronje/lachesis/src/log"
	"github.com/sirupsen/logrus"
)

type Config struct {
	HeartbeatTimeout time.Duration `mapstructure:"heartbeat"`
	TCPTimeout       time.Duration `mapstructure:"timeout"`
	CacheSize        int           `mapstructure:"cache-size"`
	SyncLimit        int           `mapstructure:"sync-limit"`
	Logger           *logrus.Logger
}

func NewConfig(heartbeat time.Duration,
	timeout time.Duration,
	cacheSize int,
	syncLimit int,
	logger *logrus.Logger) *Config {

	return &Config{
		HeartbeatTimeout: heartbeat,
		TCPTimeout:       timeout,
		CacheSize:        cacheSize,
		SyncLimit:        syncLimit,
		Logger:           logger,
	}
}

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
	}
}

func TestConfig(t *testing.T) *Config {
	config := DefaultConfig()

	config.Logger = common.NewTestLogger(t)

	return config
}
