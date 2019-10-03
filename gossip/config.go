package gossip

import (
	"math/big"

	"github.com/Fantom-foundation/go-lachesis/evm_core"
	"github.com/Fantom-foundation/go-lachesis/gossip/gasprice"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

type Config struct {
	Net     lachesis.Config
	Emitter EmitterConfig
	TxPool  evm_core.TxPoolConfig

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

// DefaultConfig returns the default configurations for the gossip service.
func DefaultConfig(network lachesis.Config) Config {
	cfg := Config{
		Net:     network,
		Emitter: DefaultEmitterConfig(),
		TxPool:  evm_core.DefaultTxPoolConfig(),
		TxIndex: true,

		GPO: gasprice.Config{
			Blocks:     20,
			Percentile: 60,
			Default:    big.NewInt(1000000000),
		},

		ForcedBroadcast: true,
	}
	if network.NetworkId == lachesis.FakeNetworkId {
		cfg.Emitter = FakeEmitterConfig()
	}
	/*if network.NetworkId == lachesis.DevNetworkId { // TODO dev network
		cfg.TxPool = evm_core.FakeTxPoolConfig()
		cfg.Emitter = FakeEmitterConfig()
	}*/
	return cfg
}
