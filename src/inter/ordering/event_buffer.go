package ordering

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// event is a inter.Event and data for ordering purpose.
type event struct {
	*inter.Event

	parents map[hash.Event]*inter.Event
}

// EventBuffer validates, bufferizes and drops() or processes() pushed() event
// if all their parents exists().
// TODO: drop incomplete events by timeout.
func EventBuffer(
	process func(*inter.Event),
	drop func(*inter.Event, error),
	exists func(hash.Event) *inter.Event) (
	push func(*inter.Event)) {

	var (
		incompletes = make(map[hash.Event]*event)
		onNewEvent  func(e *event)
	)

	onNewEvent = func(e *event) {
		reffs := newRefsValidator(e.Event)
		time := newLamportTimeValidator(e.Event)

		// fill event's parents index or hold it as incompleted
		for pHash := range e.Parents {
			if pHash.IsZero() {
				// first event of node
				if err := reffs.AddUniqueParent(e.Creator); err != nil {
					drop(e.Event, err)
					return
				}
				if err := time.AddParentTime(0); err != nil {
					drop(e.Event, err)
					return
				}
				continue
			}
			parent := e.parents[pHash]
			if parent == nil {
				parent = exists(pHash)
				if parent == nil {
					incompletes[e.Hash()] = e
					return
				}
				e.parents[pHash] = parent
			}
			if err := reffs.AddUniqueParent(parent.Creator); err != nil {
				drop(e.Event, err)
				return
			}
			if err := time.AddParentTime(parent.LamportTime); err != nil {
				drop(e.Event, err)
				return
			}
		}
		if err := reffs.CheckSelfParent(); err != nil {
			drop(e.Event, err)
			return
		}
		if err := time.CheckSequential(); err != nil {
			drop(e.Event, err)
			return
		}

		// parents OK
		process(e.Event)

		// now child events may become complete, check it again
		for hash, child := range incompletes {
			if parent, ok := child.parents[e.Hash()]; ok && parent == nil {
				child.parents[e.Hash()] = e.Event
				delete(incompletes, hash)
				onNewEvent(child)
			}
		}
	}

	push = func(e *inter.Event) {
		if exists(e.Hash()) != nil {
			drop(e, fmt.Errorf("event %s had received already", e.Hash().String()))
			return
		}

		w := &event{
			Event:   e,
			parents: make(map[hash.Event]*inter.Event, len(e.Parents)),
		}
		for hash := range e.Parents {
			w.parents[hash] = nil
		}
		onNewEvent(w)
	}

	return
}
