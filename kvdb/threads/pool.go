package threads

import (
	"runtime/debug"
	"sync"
)

const GoroutinesPerThread = 0.8

// threadPool counts threads in use
type threadPool struct {
	mu          sync.Mutex
	initialized bool
	sum         int
}

var globalPool threadPool

// init threadPool only on demand to give time to other packages
// call debug.SetMaxThreads() if they need
func (p *threadPool) init() {
	if !p.initialized {
		p.initialized = true
		p.sum = int(getMaxThreads() * GoroutinesPerThread)
	}
}

func (p *threadPool) Lock(want int) (got int, release func()) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		p.init()
	}

	if want < 1 {
		want = 0
	}

	got = min(p.sum, want)
	p.sum -= got
	release = func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.sum += got
	}

	return
}

func getMaxThreads() float64 {
	was := debug.SetMaxThreads(10000)
	debug.SetMaxThreads(was)
	return float64(was)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
