package gossip

import (
	"math/big"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/gossip/gasprice"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

type (
	// Config for the gossip service.
	Config struct {
		Net     lachesis.Config
		Emitter EmitterConfig
		TxPool  evmcore.TxPoolConfig
		StoreConfig

		// Protocol options
		TxIndex         bool // Whether to disable indexing transactions and receipts or not
		ForcedBroadcast bool

		// Gas Price Oracle options
		GPO gasprice.Config

		// Enables tracking of SHA3 preimages in the VM
		EnablePreimageRecording bool // TODO

		// Type of the EWASM interpreter ("" for default)
		EWASMInterpreter string

		// Type of the EVM interpreter ("" for default)
		EVMInterpreter string // TODO custom interpreter

		// RPCGasCap is the global gas cap for eth-call variants.
		RPCGasCap *big.Int `toml:",omitempty"`

		ExtRPCEnabled bool
	}

	// StoreConfig is a config for store db.
	StoreConfig struct {
		// Cache size for Events.
		EventsCacheSize int
		// Cache size for EventHeaderData (Epoch db).
		EventsHeadersCacheSize int
		// Cache size for Block.
		BlockCacheSize int
		// Cache size for PackInfos.
		PackInfosCacheSize int
		// Cache size for Receipts.
		ReceiptsCacheSize int
		// Cache size for TxPositions.
		TxPositionsCacheSize int
		// Cache size for EpochStats.
		EpochStatsCacheSize int
		// Cache size for LastEpochHeaders.
		LastEpochHeadersCacheSize int
		// Cache size for Stakers.
		StakersCacheSize int
		// Cache size for Delegators.
		DelegatorsCacheSize int
	}
)

// DefaultConfig returns the default configurations for the gossip service.
func DefaultConfig(network lachesis.Config) Config {
	cfg := Config{
		Net:         network,
		Emitter:     DefaultEmitterConfig(),
		TxPool:      evmcore.DefaultTxPoolConfig(),
		StoreConfig: DefaultStoreConfig(),

		TxIndex:         true,
		ForcedBroadcast: true,

		GPO: gasprice.Config{
			Blocks:     20,
			Percentile: 60,
			Default:    big.NewInt(1000000000),
		},
	}

	if network.NetworkID == lachesis.FakeNetworkID {
		cfg.Emitter = FakeEmitterConfig()
	}
	/*if network.NetworkId == lachesis.DevNetworkId { // TODO dev network
		cfg.TxPool = evmcore.FakeTxPoolConfig()
		cfg.Emitter = FakeEmitterConfig()
	}*/
	return cfg
}

// DefaultStoreConfig for product.
func DefaultStoreConfig() StoreConfig {
	return StoreConfig{
		EventsCacheSize:           300,
		EventsHeadersCacheSize:    10000,
		BlockCacheSize:            100,
		PackInfosCacheSize:        100,
		ReceiptsCacheSize:         100,
		TxPositionsCacheSize:      1000,
		EpochStatsCacheSize:       100,
		LastEpochHeadersCacheSize: 2,
		DelegatorsCacheSize:       4000,
		StakersCacheSize:          4000,
	}
}

// LiteStoreConfig is for tests or inmemory.
func LiteStoreConfig() StoreConfig {
	return StoreConfig{
		EventsCacheSize:           100,
		EventsHeadersCacheSize:    1000,
		BlockCacheSize:            100,
		PackInfosCacheSize:        100,
		ReceiptsCacheSize:         100,
		TxPositionsCacheSize:      100,
		EpochStatsCacheSize:       100,
		LastEpochHeadersCacheSize: 2,
		DelegatorsCacheSize:       400,
		StakersCacheSize:          400,
	}
}
