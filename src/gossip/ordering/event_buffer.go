package ordering

import (
	"github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-lachesis/src/event_check"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
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
		Exists  func(hash.Event) *inter.EventHeaderData
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
	if buf.callback.Exists(e.Hash()) != nil {
		buf.callback.Drop(e, peer, event_check.ErrAlreadyConnectedEvent)
		return
	}

	w := &event{
		Event: e,
		peer:  peer,
	}

	buf.pushEvent(w)
}

func (buf *EventBuffer) pushEvent(e *event) {
	// LRU is thread-safe, no need in mutex
	if buf.callback.Exists(e.Hash()) != nil {
		return
	}

	parents := make([]*inter.EventHeaderData, len(e.Parents)) // use local buffer for thread safety
	for i, p := range e.Parents {
		_, _ = buf.incompletes.Get(p) // updating the "recently used"-ness of the key
		parent := buf.callback.Exists(p)
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
	for _, childId := range buf.incompletes.Keys() {
		child, _ := buf.incompletes.Peek(childId) // without updating the "recently used"-ness of the key
		if child == nil {
			continue
		}
		for _, parent := range child.(*event).Parents {
			if parent == e.Hash() {
				buf.pushEvent(child.(*event))
			}
		}
	}

	buf.incompletes.Remove(e.Hash())
}

func (buf *EventBuffer) IsBuffered(id hash.Event) bool {
	return buf.incompletes.Contains(id) // LRU is thread-safe, no need in mutex
}

func (buf *EventBuffer) Clear() {
	buf.incompletes.Purge() // LRU is thread-safe, no need in mutex
}
