package heavycheck

type Config struct {
	MaxQueuedBatches int // the maximum number of event batches to queue up
	MaxBatch         int // Maximum number of events in an task batch (batch is divided if exceeded)
	Threads          int
}

func DefaultConfig() Config {
	return Config{
		MaxQueuedBatches: 128,
		MaxBatch:         8,
		Threads:          0,
	}
}
