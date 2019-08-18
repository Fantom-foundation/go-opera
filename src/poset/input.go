package poset

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type EventSource interface {
	HasEvent(hash.Event) bool
	GetEvent(hash.Event) *inter.Event
	GetEventHeader(hash.Event) *inter.EventHeaderData
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
		p.Fatalf("got unsaved event %s", h.String())
	}
	return &Event{
		Event: e,
	}
}

// GetEventHeader returns event header.
func (p *Poset) GetEventHeader(h hash.Event) *inter.EventHeaderData {
	return p.input.GetEventHeader(h)
}
