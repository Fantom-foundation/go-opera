package gossip

import (
	"fmt"
	"time"

	"github.com/Fantom-foundation/lachesis-base/gossip/dagfetcher"
	"github.com/Fantom-foundation/lachesis-base/gossip/dagprocessor"
	"github.com/Fantom-foundation/lachesis-base/gossip/dagstream/streamleecher"
	"github.com/Fantom-foundation/lachesis-base/gossip/dagstream/streamseeder"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/go-opera/eventcheck/heavycheck"
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/emitter"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/gossip/gasprice"
)

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

		Fetcher       dagfetcher.Config
		StreamLeecher streamleecher.Config
		StreamSeeder  streamseeder.Config
	}
	// Config for the gossip service.
	Config struct {
		Emitter emitter.Config
		TxPool  evmcore.TxPoolConfig

		TxIndex             bool // Whether to enable indexing transactions and receipts or not
		DecisiveEventsIndex bool // Whether to enable indexing events which decide blocks or not
		EventLocalTimeIndex bool // Whether to enable indexing arrival time of events or not

		// Protocol options
		Protocol ProtocolConfig

		HeavyCheck heavycheck.Config

		// Gas Price Oracle options
		GPO gasprice.Config

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
	}
	StoreCacheConfig struct {
		// Cache size for full events.
		EventsNum  int
		EventsSize uint
		// Cache size for event headers
		EventsHeadersNum int
		// Cache size for Blocks.
		BlocksNum  int
		BlocksSize uint
	}

	// StoreConfig is a config for store db.
	StoreConfig struct {
		Cache StoreCacheConfig
		// EVM is EVM store config
		EVM evmstore.StoreConfig
	}
)

// DefaultConfig returns the default configurations for the gossip service.
func DefaultConfig() Config {
	cfg := Config{
		Emitter: emitter.DefaultConfig(),
		TxPool:  evmcore.DefaultTxPoolConfig(),

		TxIndex:             true,
		DecisiveEventsIndex: false,

		HeavyCheck: heavycheck.DefaultConfig(),

		Protocol: ProtocolConfig{
			LatencyImportance:    60,
			ThroughputImportance: 40,
			MsgsSemaphoreLimit: dag.Metric{
				Num:  1000,
				Size: 30 * opt.MiB,
			},
			EventsSemaphoreLimit: dag.Metric{
				Num:  10000,
				Size: 30 * opt.MiB,
			},
			MsgsSemaphoreTimeout:    10 * time.Second,
			ProgressBroadcastPeriod: 10 * time.Second,

			Processor:     dagprocessor.DefaultConfig(),
			Fetcher:       dagfetcher.DefaultConfig(),
			StreamLeecher: streamleecher.DefaultConfig(),
			StreamSeeder:  streamseeder.DefaultConfig(),
		},

		GPO: gasprice.Config{
			Blocks:     20,
			Percentile: 60,
			MaxPrice:   gasprice.DefaultMaxPrice,
		},
	}

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
func FakeConfig(num int) Config {
	cfg := DefaultConfig()
	cfg.Emitter = emitter.FakeConfig(num)
	return cfg
}

// DefaultStoreConfig for product.
func DefaultStoreConfig() StoreConfig {
	return StoreConfig{
		Cache: StoreCacheConfig{
			EventsNum:        5000,
			EventsSize:       6 * opt.MiB,
			EventsHeadersNum: 5000,
			BlocksNum:        1000,
			BlocksSize:       512 * opt.KiB,
		},
		EVM: evmstore.DefaultStoreConfig(),
	}
}

// LiteStoreConfig is for tests or inmemory.
func LiteStoreConfig() StoreConfig {
	return StoreConfig{
		Cache: StoreCacheConfig{
			EventsNum:        500,
			EventsSize:       512 * opt.KiB,
			EventsHeadersNum: 500,
			BlocksNum:        100,
			BlocksSize:       50 * opt.KiB,
		},
		EVM: evmstore.LiteStoreConfig(),
	}
}
