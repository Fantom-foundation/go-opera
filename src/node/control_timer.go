package node

import (
	"math/rand"
	"sync"
	"time"
)

type timerFactory func(time.Duration) <-chan time.Time

type ControlTimer struct {
	timerFactory timerFactory
	tickCh       chan struct{} //sends a signal to listening process
	resetCh      chan time.Duration //receives instruction to reset the heartbeatTimer
	stopCh       chan struct{} //receives instruction to stop the heartbeatTimer
	shutdownCh   chan struct{} //receives instruction to exit Run loop
	set          bool
	Locker       sync.RWMutex
}

func NewControlTimer(timerFactory timerFactory) *ControlTimer {
	return &ControlTimer{
		timerFactory: timerFactory,
		tickCh:       make(chan struct{}),
		resetCh:      make(chan time.Duration),
		stopCh:       make(chan struct{}),
		shutdownCh:   make(chan struct{}),
	}
}

func NewRandomControlTimer() *ControlTimer {

	randomTimeout := func(min time.Duration) <-chan time.Time {
		if min == 0 {
			return nil
		}
		extra := time.Duration(rand.Int63()) % min
		return time.After(min + extra)
	}
	return NewControlTimer(randomTimeout)
}

func (c *ControlTimer) Run(init time.Duration) {

	setTimer := func(t time.Duration) <-chan time.Time {
		c.SetSet(true)
		return c.timerFactory(t)
	}

	timer := setTimer(init)
	for {
		select {
		case <-timer:
			c.tickCh <- struct{}{}
			c.SetSet(false)
		case t:= <-c.resetCh:
			timer = setTimer(t)
		case <-c.stopCh:
			timer = nil
			c.SetSet(false)
		case <-c.shutdownCh:
			c.SetSet(false)
			return
		}
	}
}

func (c *ControlTimer) Shutdown() {
	close(c.shutdownCh)
}

func (c *ControlTimer) SetSet(v bool) {
	c.Locker.Lock()
	defer c.Locker.Unlock()
	c.set = v
}

func (c *ControlTimer) GetSet() bool {
	c.Locker.RLock()
	defer c.Locker.RUnlock()
	return c.set
}
