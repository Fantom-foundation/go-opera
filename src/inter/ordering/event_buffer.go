package ordering

import (
	"fmt"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

const timeout = 2 * time.Hour

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
func EventBuffer(callback Callback) (push func(*inter.Event)) {

	var (
		incompletes = make(map[hash.Event]*event)
		onNewEvent  func(e *event)
		toDelete    = make(map[hash.Event]time.Time)
	)

	onNewEvent = func(e *event) {
		reffs := newRefsValidator(e.Event)
		timeVal := newLamportTimeValidator(e.Event)

		// fill event's parents index or hold it as incompleted
		for pHash := range e.Parents {
			if pHash.IsZero() {
				// first event of node
				if err := reffs.AddUniqueParent(e.Creator); err != nil {
					callback.Drop(e.Event, err)
					return
				}
				if err := timeVal.AddParentTime(0); err != nil {
					callback.Drop(e.Event, err)
					return
				}
				continue
			}
			parent := e.parents[pHash]
			if parent == nil {
				parent = callback.Exists(pHash)
				if parent == nil {
					h := e.Hash()
					incompletes[h] = e
					toDelete[h] = time.Now().Add(timeout)
					return
				}
				e.parents[pHash] = parent
			}
			if err := reffs.AddUniqueParent(parent.Creator); err != nil {
				callback.Drop(e.Event, err)
				return
			}
			if err := timeVal.AddParentTime(parent.LamportTime); err != nil {
				callback.Drop(e.Event, err)
				return
			}
		}
		if err := reffs.CheckSelfParent(); err != nil {
			callback.Drop(e.Event, err)
			return
		}
		if err := timeVal.CheckSequential(); err != nil {
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
				delete(toDelete, hash_)
				onNewEvent(child)
			}
		}

		// Delete by timeout
		now := time.Now()
		for hash_, t := range toDelete {
			if now.Before(t) {
				continue
			}

			delete(incompletes, hash_)
			delete(toDelete, hash_)
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
