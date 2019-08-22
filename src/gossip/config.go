package gossip

import (
	"github.com/ethereum/go-ethereum/eth/downloader"

	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
)

type Config struct {
	Net lachesis.Config

	Emitter EmitterConfig

	// Protocol options
	SyncMode downloader.SyncMode

	NoPruning       bool // Whether to disable pruning and flush everything to disk
	NoPrefetch      bool // Whether to disable prefetching and only load state on demand
	ForcedBroadcast bool
}

func DefaultConfig(network lachesis.Config) Config {
	return Config{
		Net:             network,
		ForcedBroadcast: true,
		Emitter:         DefaultEmitterConfig(),
	}
}
