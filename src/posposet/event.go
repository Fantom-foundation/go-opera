package posposet

// Event is a poset event.
type Event struct {
	Creator Address
	Parents EventHashes

	hash    EventHash
	parents map[EventHash]*Event
}

// Hash calcs hash of event.
func (e *Event) Hash() EventHash {
	return EventHashOf(e)
}
