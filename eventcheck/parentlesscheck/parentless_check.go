package parentlesscheck

import (
	"github.com/Fantom-foundation/lachesis-base/eventcheck/epochcheck"
	"github.com/Fantom-foundation/lachesis-base/eventcheck/queuedcheck"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
)

type Checker struct {
	callback Callback
}

type LightCheck func(dag.Event) error

type HeavyCheck interface {
	Enqueue(tasks []queuedcheck.EventTask, onValidated func([]queuedcheck.EventTask)) error
}

type Callback struct {
	// FilterInterested returns only event which may be requested.
	OnlyInterested func(ids hash.Events) hash.Events

	HeavyCheck HeavyCheck
	LightCheck LightCheck
}

func New(callback Callback) *Checker {
	return &Checker{
		callback: callback,
	}
}

// Enqueue tries to fill gaps the fetcher's future import queue.
func (c *Checker) Enqueue(tasks []queuedcheck.EventTask, checked func([]queuedcheck.EventTask)) error {
	// Run light checks right away
	passed := make([]queuedcheck.EventTask, 0, len(tasks))
	for _, t := range tasks {
		// Filter already known events
		if len(c.callback.OnlyInterested(hash.Events{t.Event().ID()})) == 0 {
			t.SetResult(epochcheck.ErrNotRelevant)
			checked([]queuedcheck.EventTask{t})
			continue
		}

		err := c.callback.LightCheck(t.Event())
		if err != nil {
			t.SetResult(err)
			checked([]queuedcheck.EventTask{t})
			continue
		}
		passed = append(passed, t)
	}

	// Run heavy check in parallel
	return c.callback.HeavyCheck.Enqueue(passed, checked)
}
