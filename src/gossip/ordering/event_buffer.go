package ordering

import (
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/event_check"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

const expiration = 1 * time.Hour

type (
	// event is a inter.Event and data for ordering purpose.
	event struct {
		*inter.Event

		peer    string
		parents map[hash.Event]*inter.Event
		expired time.Time
	}

	// Callback is a set of EventBuffer()'s args.
	Callback struct {
		Process func(e *inter.Event) error
		Drop    func(e *inter.Event, peer string, err error)
		Exists  func(hash.Event) *inter.Event
		Check   func(e *inter.Event, parents []*inter.EventHeaderData) error
	}
)

type IsBufferedFn func(hash.Event) bool
type PushEventFn func(e *inter.Event, peer string)

// EventBuffer validates, bufferizes and drops() or processes() pushed() event
// if all their parents exists().
func EventBuffer(callback Callback) (push PushEventFn, downloaded IsBufferedFn) {

	var (
		incompletes = make(map[hash.Event]*event)
		lastGC      = time.Now()
		onNewEvent  func(e *event, peer string)
	)

	onNewEvent = func(e *event, peer string) {
		for _, p := range e.Parents {
			parent := e.parents[p]
			if parent == nil {
				parent = callback.Exists(p)
				if parent == nil {
					h := e.Hash()
					incompletes[h] = e
					return
				}
				e.parents[p] = parent
			}
		}

		// validate
		if callback.Check != nil {
			parentHeaders := make([]*inter.EventHeaderData, len(e.Parents))
			for i, p := range e.Parents {
				parentHeaders[i] = &e.parents[p].EventHeaderData
			}
			err := callback.Check(e.Event, parentHeaders)
			if err != nil {
				callback.Drop(e.Event, peer, err)
				return
			}
		}

		// process
		err := callback.Process(e.Event)
		if err != nil {
			callback.Drop(e.Event, peer, err)
			return
		}

		// now child events may become complete, check it again
		for hash_, child := range incompletes {
			if parent, ok := child.parents[e.Hash()]; ok && parent == nil {
				child.parents[e.Hash()] = e.Event
				delete(incompletes, hash_)
				onNewEvent(child, child.peer)
			}
		}

	}

	push = func(e *inter.Event, peer string) {
		if callback.Exists(e.Hash()) != nil {
			callback.Drop(e, peer, event_check.ErrAlreadyConnectedEvent)
			return
		}

		w := &event{
			Event:   e,
			peer:    peer,
			parents: make(map[hash.Event]*inter.Event, len(e.Parents)),
			expired: time.Now().Add(expiration),
		}
		for _, parentHash := range e.Parents {
			w.parents[parentHash] = nil
		}
		onNewEvent(w, peer)

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

	downloaded = func(id hash.Event) bool {
		_, ok := incompletes[id]
		return ok
	}

	return
}
