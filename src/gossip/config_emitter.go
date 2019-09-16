package gossip

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type EmitterConfig struct {
	Emitbase common.Address

	MinEmitInterval time.Duration // minimum event emission interval
	MaxEmitInterval time.Duration // maximum event emission interval
}

func DefaultEmitterConfig() EmitterConfig {
	return EmitterConfig{
		MinEmitInterval: 1 * time.Second,
		MaxEmitInterval: 60 * time.Second,
	}
}
