package metrics

import (
	"time"
)

type throttler struct {
	Period  uint
	Timeout time.Duration
	count   uint
}

func (t *throttler) Do() {
	if t.Period == 0 {
		return
	}

	t.count++
	if t.count%t.Period == 0 {
		time.Sleep(t.Timeout)
	}
}
