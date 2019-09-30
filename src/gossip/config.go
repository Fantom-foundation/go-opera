package gossip

import (
	"math/big"

	"github.com/ethereum/go-ethereum/eth/downloader"

	"github.com/Fantom-foundation/go-lachesis/src/evm_core"
	"github.com/Fantom-foundation/go-lachesis/src/gossip/gasprice"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
)

type Config struct {
	Net     lachesis.Config
	Emitter EmitterConfig
	TxPool  evm_core.TxPoolConfig

	// Protocol options
	SyncMode downloader.SyncMode

	NoPruning       bool // Whether to disable pruning and flush everything to disk
	NoPrefetch      bool // Whether to disable prefetching and only load state on demand
	ForcedBroadcast bool

	// Gas Price Oracle options
	GPO gasprice.Config

	// Enables tracking of SHA3 preimages in the VM
	EnablePreimageRecording bool

	// Type of the EWASM interpreter ("" for default)
	EWASMInterpreter string

	// Type of the EVM interpreter ("" for default)
	EVMInterpreter string

	// Miscellaneous options
	DocRoot string `toml:"-"`

	// RPCGasCap is the global gas cap for eth-call variants.
	RPCGasCap *big.Int `toml:",omitempty"`

	ExtRPCEnabled bool
}

// DefaultConfig returns the default configurations for the gossip service.
func DefaultConfig(network lachesis.Config) Config {
	return Config{
		Net:     network,
		Emitter: DefaultEmitterConfig(),
		TxPool:  evm_core.DefaultTxPoolConfig(),

		GPO: gasprice.Config{
			Blocks:     20,
			Percentile: 60,
			Default:    big.NewInt(1000000000),
		},

		ForcedBroadcast: true,
	}
}

// FakeConfig returns the fake configurations for the gossip service.
func FakeConfig() Config {
	network := lachesis.FakeNetConfig(3)
	config := DefaultConfig(network)
	config.TxPool = evm_core.FakeTxPoolConfig()

	return config
}
