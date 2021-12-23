package gossip

import (
	"sync"
	"time"
)

type PeriodicFlusherCallaback struct {
	busy         func() bool
	commitNeeded func() bool
	commit       func()
}

// PeriodicFlusher periodically commits the Store if isCommitNeeded returns true
type PeriodicFlusher struct {
	period   time.Duration
	callback PeriodicFlusherCallaback

	wg   sync.WaitGroup
	quit chan struct{}
}

func (c *PeriodicFlusher) loop() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.period)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if !c.callback.busy() && c.callback.commitNeeded() {
				c.callback.commit()
			}
		case <-c.quit:
			return
		}
	}
}

func (c *PeriodicFlusher) Start() {
	c.wg.Add(1)
	go c.loop()
}

func (c *PeriodicFlusher) Stop() {
	close(c.quit)
	c.wg.Wait()
}
