package ordering

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type (
	// event is a inter.Event and data for ordering purpose.
	event struct {
		*inter.Event

		parents map[hash.Event]*inter.Event
	}

	// Callback is a set of EventBuffer()'s args.
	Callback struct {
		Process func(*inter.Event)
		Drop    func(*inter.Event, error)
		Exists  func(hash.Event) *inter.Event
	}
)

// EventBuffer validates, bufferizes and drops() or processes() pushed() event
// if all their parents exists().
// TODO: drop incomplete events by timeout.
func EventBuffer(callback Callback) (push func(*inter.Event)) {

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
					callback.Drop(e.Event, err)
					return
				}
				if err := time.AddParentTime(0); err != nil {
					callback.Drop(e.Event, err)
					return
				}
				continue
			}
			parent := e.parents[pHash]
			if parent == nil {
				parent = callback.Exists(pHash)
				if parent == nil {
					incompletes[e.Hash()] = e
					return
				}
				e.parents[pHash] = parent
			}
			if err := reffs.AddUniqueParent(parent.Creator); err != nil {
				callback.Drop(e.Event, err)
				return
			}
			if err := time.AddParentTime(parent.LamportTime); err != nil {
				callback.Drop(e.Event, err)
				return
			}
		}
		if err := reffs.CheckSelfParent(); err != nil {
			callback.Drop(e.Event, err)
			return
		}
		if err := time.CheckSequential(); err != nil {
			callback.Drop(e.Event, err)
			return
		}

		// parents OK
		callback.Process(e.Event)

		// now child events may become complete, check it again
		for hash_, child := range incompletes {
			if parent, ok := child.parents[e.Hash()]; ok && parent == nil {
				child.parents[e.Hash()] = e.Event
				delete(incompletes, hash_)
				onNewEvent(child)
			}
		}
	}

	push = func(e *inter.Event) {
		if callback.Exists(e.Hash()) != nil {
			callback.Drop(e, fmt.Errorf("event %s had received already", e.Hash().String()))
			return
		}

		w := &event{
			Event:   e,
			parents: make(map[hash.Event]*inter.Event, len(e.Parents)),
		}
		for parentHash := range e.Parents {
			w.parents[parentHash] = nil
		}
		onNewEvent(w)
	}

	return
}
