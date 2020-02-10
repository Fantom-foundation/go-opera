package poset

import (
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// EventSource is a callback for getting events from an external storage.
type EventSource interface {
	HasEvent(hash.Event) bool
	GetEvent(hash.Event) *inter.Event
	GetEventHeader(idx.Epoch, hash.Event) *inter.EventHeaderData
}

/*
 * Poset's methods:
 */

// GetEvent returns event.
func (p *Poset) GetEvent(h hash.Event) *Event {
	e := p.input.GetEvent(h)
	if e == nil {
		p.Log.Crit("Got unsaved event", "event", h.String())
	}
	return &Event{
		Event: e,
	}
}

// GetEventHeader returns event header.
func (p *Poset) GetEventHeader(epoch idx.Epoch, h hash.Event) *inter.EventHeaderData {
	return p.input.GetEventHeader(epoch, h)
}
