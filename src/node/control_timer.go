package node

import (
	"math/rand"
	"sync"
	"time"
)

type timerFactory func(time.Duration) <-chan time.Time

// ControlTimer struct that controls timing events in the node
type ControlTimer struct {
	timerFactory timerFactory
	tickCh       chan struct{}      //sends a signal to listening process
	resetCh      chan time.Duration //receives instruction to reset the heartbeatTimer
	stopCh       chan struct{}      //receives instruction to stop the heartbeatTimer
	shutdownCh   chan struct{}      //receives instruction to exit Run loop
	set          bool
	Locker       sync.RWMutex
}

// NewControlTimer creates a new control timer struct
func NewControlTimer(timerFactory timerFactory) *ControlTimer {
	return &ControlTimer{
		timerFactory: timerFactory,
		tickCh:       make(chan struct{}),
		resetCh:      make(chan time.Duration),
		stopCh:       make(chan struct{}),
		shutdownCh:   make(chan struct{}),
	}
}

// NewRandomControlTimer creates a random time controller with no defaults set
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

// Run handles all the time based events in the background
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
		case t := <-c.resetCh:
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

// Shutdown the control timer
func (c *ControlTimer) Shutdown() {
	close(c.shutdownCh)
}

// SetSet sets the set value for the set
func (c *ControlTimer) SetSet(v bool) {
	c.Locker.Lock()
	defer c.Locker.Unlock()
	c.set = v
}

// GetSet retrieves the set
func (c *ControlTimer) GetSet() bool {
	c.Locker.RLock()
	defer c.Locker.RUnlock()
	return c.set
}
