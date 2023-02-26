package threads

import (
	"runtime/debug"
	"sync"
)

const GoroutinesPerThread = 0.8

// threadPool counts threads in use
type ThreadPool struct {
	mu   sync.Mutex
	cap  int
	left int
}

var GlobalPool ThreadPool

// init ThreadPool only on demand to give time to other packages
// call debug.SetMaxThreads() if they need
func (p *ThreadPool) init() {
	if p.cap == 0 {
		p.cap = int(getMaxThreads() * GoroutinesPerThread)
		p.left = p.cap
	}
}

// Capacity of pool
func (p *ThreadPool) Cap() int {
	if p.cap == 0 {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.init()
	}
	return p.cap
}

func (p *ThreadPool) Lock(want int) (got int, release func(count int)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.init()

	if want < 1 {
		want = 0
	}

	got = min(p.left, want)
	p.left -= got

	release = func(count int) {
		p.mu.Lock()
		defer p.mu.Unlock()

		if 0 > count || count > got {
			count = got
		}

		got -= count
		p.left += count
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
