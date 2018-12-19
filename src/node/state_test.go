package node

import (
	"sync"
	"testing"
	"time"
)

func TestChangeNodeState(t *testing.T) {
	limit := 10

	wg := sync.WaitGroup{}
	wg.Add(limit * 2) // 10 write, 10 read

	ns := newNodeState2()

	setInstance := func(state state) {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			ns.setState(state)
		}
	}

	getInstance := func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			s := ns.getState()
			_ = s
		}
	}

	go func() {
		for i := 0; i < limit; i++ {
			go setInstance(Gossiping)
		}
	}()

	go func() {
		for i := 0; i < limit; i++ {
			go getInstance()
		}
	}()

	wg.Wait()
}

func TestConcurrentGoFuncs(t *testing.T) {
	ns := newNodeState2()

	f := func() {
		time.Sleep(time.Microsecond * 10)
	}

	wg := sync.WaitGroup{}
	wg.Add(2000)

	go func() {
		for i := 0; i < 1000; i++ {
			go func() {
				defer wg.Done()
				ns.goFunc(f)
			}()
		}
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			go func() {
				defer wg.Done()
				ns.waitRoutines()
			}()
		}
	}()

	wg.Wait()
}
