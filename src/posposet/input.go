package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type EventSource interface {
	HasEvent(hash.Event) bool
	GetEvent(hash.Event) *inter.Event
}

/*
 * Poset's methods:
 */

// HasEvent returns true if event exists.
func (p *Poset) HasEvent(h hash.Event) bool {
	return p.input.HasEvent(h)
}

// GetEvent returns event.
func (p *Poset) GetEvent(h hash.Event) *Event {
	e := p.input.GetEvent(h)
	if e == nil {
		p.Fatal("got unsaved event")
	}
	return &Event{
		Event: e,
	}
}
