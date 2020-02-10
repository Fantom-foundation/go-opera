package ordering

import (
	"github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-lachesis/eventcheck"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
)

type (
	// event is a inter.Event and data for ordering purpose.
	event struct {
		*inter.Event

		peer string
	}

	// Callback is a set of EventBuffer()'s args.
	Callback struct {
		Process func(e *inter.Event) error
		Drop    func(e *inter.Event, peer string, err error)
		Get     func(hash.Event) *inter.EventHeaderData
		Exists  func(hash.Event) bool
		Check   func(e *inter.Event, parents []*inter.EventHeaderData) error
	}
)

type EventBuffer struct {
	incompletes *lru.Cache // event hash -> event
	callback    Callback
}

func New(buffSize int, callback Callback) *EventBuffer {
	incompletes, _ := lru.New(buffSize)
	return &EventBuffer{
		incompletes: incompletes,
		callback:    callback,
	}
}

func (buf *EventBuffer) PushEvent(e *inter.Event, peer string) {
	w := &event{
		Event: e,
		peer:  peer,
	}

	buf.pushEvent(w, buf.getIncompleteEventsList(), true)
}

func (buf *EventBuffer) getIncompleteEventsList() []*event {
	res := make([]*event, 0, buf.incompletes.Len())
	for _, childID := range buf.incompletes.Keys() {
		child, _ := buf.incompletes.Peek(childID)
		if child == nil {
			continue
		}
		res = append(res, child.(*event))
	}
	return res
}

func (buf *EventBuffer) pushEvent(e *event, incompleteEventsList []*event, strict bool) {
	// LRU is thread-safe, no need in mutex
	if buf.callback.Exists(e.Hash()) {
		if strict {
			buf.callback.Drop(e.Event, e.peer, eventcheck.ErrAlreadyConnectedEvent)
		}
		return
	}

	parents := make([]*inter.EventHeaderData, len(e.Parents)) // use local buffer for thread safety
	for i, p := range e.Parents {
		_, _ = buf.incompletes.Get(p) // updating the "recently used"-ness of the key
		parent := buf.callback.Get(p)
		if parent == nil {
			buf.incompletes.Add(e.Hash(), e)
			return
		}
		parents[i] = parent
	}

	// validate
	if buf.callback.Check != nil {
		err := buf.callback.Check(e.Event, parents)
		if err != nil {
			buf.callback.Drop(e.Event, e.peer, err)
			return
		}
	}

	// process
	err := buf.callback.Process(e.Event)
	if err != nil {
		buf.callback.Drop(e.Event, e.peer, err)
		return
	}

	// now child events may become complete, check it again
	eHash := e.Hash()
	buf.incompletes.Remove(eHash)
	for _, child := range incompleteEventsList {
		for _, parent := range child.Parents {
			if parent == eHash {
				buf.pushEvent(child, incompleteEventsList, false)
			}
		}
	}
}

func (buf *EventBuffer) IsBuffered(id hash.Event) bool {
	return buf.incompletes.Contains(id) // LRU is thread-safe, no need in mutex
}

func (buf *EventBuffer) Clear() {
	buf.incompletes.Purge() // LRU is thread-safe, no need in mutex
}
