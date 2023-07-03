package threads

import (
	"math"
	"runtime/debug"
	"sync/atomic"
)

const GoroutinesPerThread = 0.8

// threadPool counts threads in use
type ThreadPool struct {
	cap  int32
	left int32
}

var GlobalPool ThreadPool

// init ThreadPool only on demand to give time to other packages
// call debug.SetMaxThreads() if they need.
// Note: in datarace case the other participants will see no free threads but
// enough capacity. It's ok.
func (p *ThreadPool) init() {
	initialized := !atomic.CompareAndSwapInt32(&p.cap, 0, math.MaxInt32)
	if initialized {
		return
	}
	cap := int32(getMaxThreads() * GoroutinesPerThread)
	atomic.StoreInt32(&p.left, cap)
	atomic.StoreInt32(&p.cap, cap)
}

// Capacity of pool.
// Note: first call may return greater value than nexts. Don't cache it.
func (p *ThreadPool) Cap() int {
	p.init()
	cap := atomic.LoadInt32(&p.cap)
	return int(cap)
}

func (p *ThreadPool) Lock(want int) (got int, release func(count int)) {
	p.init()

	if want < 0 {
		want = 0
	}

	left := atomic.AddInt32(&p.left, -int32(want))
	got = want
	if left < 0 {
		if left < -int32(got) {
			left = -int32(got)
		}
		atomic.AddInt32(&p.left, -left)
		got += int(left)
	}

	release = func(count int) {
		if 0 > count || count > got {
			count = got
		}

		got -= count
		atomic.AddInt32(&p.left, int32(count))
	}

	return
}

func getMaxThreads() float64 {
	was := debug.SetMaxThreads(10000)
	debug.SetMaxThreads(was)
	return float64(was)
}
