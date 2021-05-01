package gossip

import (
	"fmt"
	"time"

	"github.com/Fantom-foundation/lachesis-base/gossip/dagprocessor"
	"github.com/Fantom-foundation/lachesis-base/gossip/dagstream/streamleecher"
	"github.com/Fantom-foundation/lachesis-base/gossip/dagstream/streamseeder"
	"github.com/Fantom-foundation/lachesis-base/gossip/itemsfetcher"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/go-opera/eventcheck/heavycheck"
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/verwatcher"
	"github.com/Fantom-foundation/go-opera/gossip/emitter"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/gossip/gasprice"
)

const nominalSize uint = 1

type (
	// ProtocolConfig is config for p2p protocol
	ProtocolConfig struct {
		// 0/M means "optimize only for throughput", N/0 means "optimize only for latency", N/M is a balanced mode

		LatencyImportance    int
		ThroughputImportance int

		EventsSemaphoreLimit dag.Metric
		MsgsSemaphoreLimit   dag.Metric
		MsgsSemaphoreTimeout time.Duration

		ProgressBroadcastPeriod time.Duration

		Processor dagprocessor.Config

		DagFetcher    itemsfetcher.Config
		TxFetcher     itemsfetcher.Config
		StreamLeecher streamleecher.Config
		StreamSeeder  streamseeder.Config

		MaxInitialTxHashesSend   int
		MaxRandomTxHashesSend    int
		RandomTxHashesSendPeriod time.Duration

		PeerCache PeerCacheConfig
	}
	// Config for the gossip service.
	Config struct {
		Emitter emitter.Config
		TxPool  evmcore.TxPoolConfig

		TxIndex bool // Whether to enable indexing transactions and receipts or not

		// Protocol options
		Protocol ProtocolConfig

		HeavyCheck heavycheck.Config

		// Gas Price Oracle options
		GPO gasprice.Config

		VersionWatcher verwatcher.Config

		// Enables tracking of SHA3 preimages in the VM
		EnablePreimageRecording bool // TODO

		// Type of the EWASM interpreter ("" for default)
		EWASMInterpreter string

		// Type of the EVM interpreter ("" for default)
		EVMInterpreter string // TODO custom interpreter

		// RPCGasCap is the global gas cap for eth-call variants.
		RPCGasCap uint64 `toml:",omitempty"`

		// RPCTxFeeCap is the global transaction fee(price * gaslimit) cap for
		// send-transction variants. The unit is ether.
		RPCTxFeeCap float64 `toml:",omitempty"`

		ExtRPCEnabled bool

		RPCLogsBloom bool
	}
	StoreCacheConfig struct {
		// Cache size for full events.
		EventsNum  int
		EventsSize uint
		// Cache size for full blocks.
		BlocksNum  int
		BlocksSize uint
	}

	// StoreConfig is a config for store db.
	StoreConfig struct {
		Cache StoreCacheConfig
		// EVM is EVM store config
		EVM                 evmstore.StoreConfig
		MaxNonFlushedSize   int
		MaxNonFlushedPeriod time.Duration
	}
)

type PeerCacheConfig struct {
	MaxKnownTxs    int // Maximum transactions hashes to keep in the known list (prevent DOS)
	MaxKnownEvents int // Maximum event hashes to keep in the known list (prevent DOS)
	// MaxQueuedItems is the maximum number of items to queue up before
	// dropping broadcasts. This is a sensitive number as a transaction list might
	// contain a single transaction, or thousands.
	MaxQueuedItems idx.Event
	MaxQueuedSize  uint64
}

// DefaultConfig returns the default configurations for the gossip service.
func DefaultConfig(scale cachescale.Func) Config {
	cfg := Config{
		Emitter: emitter.DefaultConfig(),
		TxPool:  evmcore.DefaultTxPoolConfig(),

		TxIndex: true,

		HeavyCheck: heavycheck.DefaultConfig(),

		Protocol: ProtocolConfig{
			LatencyImportance:    60,
			ThroughputImportance: 40,
			MsgsSemaphoreLimit: dag.Metric{
				Num:  scale.Events(1000),
				Size: scale.U64(30 * opt.MiB),
			},
			EventsSemaphoreLimit: dag.Metric{
				Num:  scale.Events(10000),
				Size: scale.U64(30 * opt.MiB),
			},
			MsgsSemaphoreTimeout:    10 * time.Second,
			ProgressBroadcastPeriod: 10 * time.Second,

			Processor: dagprocessor.DefaultConfig(scale),
			DagFetcher: itemsfetcher.Config{
				ForgetTimeout:       1 * time.Minute,
				ArriveTimeout:       1000 * time.Millisecond,
				GatherSlack:         100 * time.Millisecond,
				HashLimit:           20000,
				MaxBatch:            scale.I(512),
				MaxQueuedBatches:    scale.I(32),
				MaxParallelRequests: 192,
			},
			TxFetcher: itemsfetcher.Config{
				ForgetTimeout:       1 * time.Minute,
				ArriveTimeout:       1000 * time.Millisecond,
				GatherSlack:         100 * time.Millisecond,
				HashLimit:           10000,
				MaxBatch:            scale.I(512),
				MaxQueuedBatches:    scale.I(32),
				MaxParallelRequests: 64,
			},
			StreamLeecher:            streamleecher.DefaultConfig(),
			StreamSeeder:             streamseeder.DefaultConfig(scale),
			MaxInitialTxHashesSend:   20000,
			MaxRandomTxHashesSend:    128,
			RandomTxHashesSendPeriod: 20 * time.Second,
			PeerCache:                DefaultPeerCacheConfig(scale),
		},

		GPO: gasprice.Config{
			Blocks:     20,
			Percentile: 60,
			MaxPrice:   gasprice.DefaultMaxPrice,
		},

		VersionWatcher: verwatcher.Config{
			ShutDownIfNotUpgraded:     false,
			WarningIfNotUpgradedEvery: 5 * time.Second,
		},
		RPCLogsBloom: true,
	}
	cfg.Protocol.Processor.EventsBufferLimit.Num = idx.Event(cfg.Protocol.StreamLeecher.Session.ParallelChunksDownload)*cfg.Protocol.StreamLeecher.Session.DefaultChunkSize.Num + softLimitItems
	cfg.Protocol.Processor.EventsBufferLimit.Size = uint64(cfg.Protocol.StreamLeecher.Session.ParallelChunksDownload)*cfg.Protocol.StreamLeecher.Session.DefaultChunkSize.Size + 8*opt.MiB
	cfg.Protocol.StreamLeecher.MaxSessionRestart = 4 * time.Minute
	cfg.Protocol.DagFetcher.ArriveTimeout = 4 * time.Second
	cfg.Protocol.DagFetcher.HashLimit = 10000
	cfg.Protocol.TxFetcher.HashLimit = 10000

	return cfg
}

func (c *Config) Validate() error {
	if c.Protocol.StreamLeecher.Session.DefaultChunkSize.Num > hardLimitItems-1 {
		return fmt.Errorf("DefaultChunkSize.Num has to be at not greater than %d", hardLimitItems-1)
	}
	if c.Protocol.StreamLeecher.Session.DefaultChunkSize.Size > protocolMaxMsgSize/2 {
		return fmt.Errorf("DefaultChunkSize.Num has to be at not greater than %d", protocolMaxMsgSize/2)
	}
	if c.Protocol.EventsSemaphoreLimit.Num < 2*c.Protocol.StreamLeecher.Session.DefaultChunkSize.Num ||
		c.Protocol.EventsSemaphoreLimit.Size < 2*c.Protocol.StreamLeecher.Session.DefaultChunkSize.Size {
		return fmt.Errorf("EventsSemaphoreLimit has to be at least 2 times greater than %s (DefaultChunkSize)", c.Protocol.StreamLeecher.Session.DefaultChunkSize.String())
	}
	if c.Protocol.EventsSemaphoreLimit.Num < 2*c.Protocol.Processor.EventsBufferLimit.Num ||
		c.Protocol.EventsSemaphoreLimit.Size < 2*c.Protocol.Processor.EventsBufferLimit.Size {
		return fmt.Errorf("EventsSemaphoreLimit has to be at least 2 times greater than %s (EventsBufferLimit)", c.Protocol.Processor.EventsBufferLimit.String())
	}
	if c.Protocol.EventsSemaphoreLimit.Size < 2*protocolMaxMsgSize {
		return fmt.Errorf("EventsSemaphoreLimit.Size has to be at least %d", 2*protocolMaxMsgSize)
	}
	if c.Protocol.MsgsSemaphoreLimit.Size < protocolMaxMsgSize {
		return fmt.Errorf("MsgsSemaphoreLimit.Size has to be at least %d", protocolMaxMsgSize)
	}
	if c.Protocol.Processor.EventsBufferLimit.Size < protocolMaxMsgSize {
		return fmt.Errorf("EventsBufferLimit.Size has to be at least %d", protocolMaxMsgSize)
	}

	return nil
}

// FakeConfig returns the default configurations for the gossip service in fakenet.
func FakeConfig(num int, scale cachescale.Func) Config {
	cfg := DefaultConfig(scale)
	cfg.Emitter = emitter.FakeConfig(num)
	return cfg
}

// DefaultStoreConfig for product.
func DefaultStoreConfig(scale cachescale.Func) StoreConfig {
	return StoreConfig{
		Cache: StoreCacheConfig{
			EventsNum:  scale.I(5000),
			EventsSize: scale.U(6 * opt.MiB),
			BlocksNum:  scale.I(5000),
			BlocksSize: scale.U(512 * opt.KiB),
		},
		EVM:                 evmstore.DefaultStoreConfig(scale),
		MaxNonFlushedSize:   22 * opt.MiB,
		MaxNonFlushedPeriod: 30 * time.Minute,
	}
}

// LiteStoreConfig is for tests or inmemory.
func LiteStoreConfig() StoreConfig {
	return StoreConfig{
		Cache: StoreCacheConfig{
			EventsNum:  500,
			EventsSize: 512 * opt.KiB,
			BlocksNum:  100,
			BlocksSize: 50 * opt.KiB,
		},
		EVM:                 evmstore.LiteStoreConfig(),
		MaxNonFlushedSize:   800 * opt.KiB,
		MaxNonFlushedPeriod: 30 * time.Minute,
	}
}

func DefaultPeerCacheConfig(scale cachescale.Func) PeerCacheConfig {
	return PeerCacheConfig{
		MaxKnownTxs:    scale.I(24576),
		MaxKnownEvents: scale.I(24576),
		MaxQueuedItems: scale.Events(4096),
		MaxQueuedSize:  scale.U64(protocolMaxMsgSize + 1024),
	}
}
