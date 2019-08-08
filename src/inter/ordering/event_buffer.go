package ordering

import (
	"fmt"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

const expiration = 1 * time.Hour

type (
	// event is a inter.Event and data for ordering purpose.
	event struct {
		*inter.Event

		parents map[hash.Event]*inter.Event
		expired time.Time
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
		lastGC      = time.Now()
		onNewEvent  func(e *event)
	)

	onNewEvent = func(e *event) {
		reffs := newRefsValidator(e.Event)
		ltime := newLamportTimeValidator(e.Event)

		// fill event's parents index or hold it as incompleted
		if e.SelfParent() == nil {
			// first event of node
			if err := reffs.AddUniqueParent(e.Creator); err != nil {
				callback.Drop(e.Event, err)
				return
			}
			if err := ltime.AddParentTime(0); err != nil {
				callback.Drop(e.Event, err)
				return
			}
		}

		for _, pHash := range e.Parents {
			parent := e.parents[pHash]
			if parent == nil {
				parent = callback.Exists(pHash)
				if parent == nil {
					h := e.Hash()
					incompletes[h] = e
					return
				}
				e.parents[pHash] = parent
			}
			if err := reffs.AddUniqueParent(parent.Creator); err != nil {
				callback.Drop(e.Event, err)
				return
			}
			if parent.Creator == e.Creator && !e.IsSelfParent(pHash) {
				callback.Drop(e.Event, fmt.Errorf("invalid SelfParent"))
				return
			}
			if err := ltime.AddParentTime(parent.Lamport); err != nil {
				callback.Drop(e.Event, err)
				return
			}
		}
		if err := reffs.CheckSelfParent(); err != nil {
			callback.Drop(e.Event, err)
			return
		}
		if err := ltime.CheckSequential(); err != nil {
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
			expired: time.Now().Add(expiration),
		}
		for _, parentHash := range e.Parents {
			w.parents[parentHash] = nil
		}
		onNewEvent(w)

		// GC
		if time.Now().Add(-expiration / 4).Before(lastGC) {
			return
		}
		lastGC = time.Now()
		limit := time.Now().Add(-expiration)
		for h, e := range incompletes {
			if e.expired.Before(limit) {
				delete(incompletes, h)
			}
		}
	}

	return
}
