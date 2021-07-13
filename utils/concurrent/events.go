package concurrent

import (
	"sync"

	"github.com/Fantom-foundation/lachesis-base/hash"
)

type EventsSet struct {
	sync.RWMutex
	hash.EventsSet
}

func WrapEventsSet(v hash.EventsSet) *EventsSet {
	return &EventsSet{
		RWMutex:   sync.RWMutex{},
		EventsSet: v,
	}
}
