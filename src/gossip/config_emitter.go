package cfg_emitter

import "time"

type Config struct {
	MinEmitInterval time.Duration // minimum event emission interval
	MaxEmitInterval time.Duration // maximum event emission interval
}
