package posposet

// Event is a poset event.
type Event struct {
	Index   uint64
	Creator Address
	Parents EventHashes

	creator *Node
	parents []*Event
}

// Hash calcs hash of event.
func (e *Event) Hash() EventHash {
	return EventHashOf(e)
}
