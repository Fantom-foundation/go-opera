package heavycheck

type Config struct {
	MaxQueuedTasks int // the maximum number of tasks to queue up
	Threads        int
}

func DefaultConfig() Config {
	return Config{
		MaxQueuedTasks: 1024,
		Threads:        0,
	}
}
