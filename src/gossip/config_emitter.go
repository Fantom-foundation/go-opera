package gossip

import (
	"time"
)

type EmitterConfig struct {
	MinEmitInterval time.Duration // minimum event emission interval
	MaxEmitInterval time.Duration // maximum event emission interval
}

func DefaultEmitterConfig() EmitterConfig {
	return EmitterConfig{
		MinEmitInterval: 1 * time.Second,
		MaxEmitInterval: 60 * time.Second,
	}
}
