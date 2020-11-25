package parentlesscheck

import (
	"github.com/Fantom-foundation/lachesis-base/eventcheck/epochcheck"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
)

type Checker struct {
	callback Callback
}

type LightCheck func(dag.Event) error

type HeavyCheck interface {
	Enqueue(events dag.Events, onValidated func(dag.Events, []error)) error
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
func (c *Checker) Enqueue(inEvents dag.Events, checked func(ee dag.Events, errs []error)) error {
	// Run light checks right away
	passed := make(dag.Events, 0, len(inEvents))
	for _, e := range inEvents {
		// Filter already known events
		if len(c.callback.OnlyInterested(hash.Events{e.ID()})) == 0 {
			checked(dag.Events{e}, []error{epochcheck.ErrNotRelevant})
			continue
		}

		err := c.callback.LightCheck(e)
		if err != nil {
			checked(dag.Events{e}, []error{err})
			continue
		}
		passed = append(passed, e)
	}

	// Run heavy check in parallel
	return c.callback.HeavyCheck.Enqueue(passed, checked)
}
