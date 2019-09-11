package gossip

import (
	"github.com/ethereum/go-ethereum/eth/downloader"

	"github.com/Fantom-foundation/go-lachesis/src/evm_core"
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

	ExtRPCEnabled bool
}

// DefaultConfig returns the default configurations for the gossip service.
func DefaultConfig(network lachesis.Config) Config {
	return Config{
		Net:     network,
		Emitter: DefaultEmitterConfig(),
		TxPool:  evm_core.DefaultTxPoolConfig(),

		ForcedBroadcast: true,
	}
}

// FakeConfig returns the fake configurations for the gossip service.
func FakeConfig() Config {
	network := lachesis.FakeNetConfig(3)

	return Config{
		Net:     network,
		Emitter: DefaultEmitterConfig(),
		TxPool:  evm_core.FakeTxPoolConfig(),

		ForcedBroadcast: true,
	}
}
