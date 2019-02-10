package posposet

// TODO: make FlagTable internal

// TODO: cache PoS-stake at FlagTable.eventKnowsRoots

// FlagTable stores the reachability of each top event to the roots.
// It helps to select root without using path searching algorithms.
// Zero-hash is a self-parent root.
type FlagTable struct {
	//lastNodeEvent   map[common.Address]EventHash
	eventKnowsRoots map[EventHash]EventHashes
}

func NewFlagTable() *FlagTable {
	return &FlagTable{
		//lastNodeEvent:   make(map[common.Address]EventHash),
		eventKnowsRoots: make(map[EventHash]EventHashes),
	}
}

func (ft *FlagTable) KnownRootsOf(e *Event, parent EventHash) *EventHashes {
	if parent.IsZero() {
		// self-root
		return &EventHashes{
			index: map[EventHash]struct{}{
				e.Hash(): struct{}{},
			},
		}
	}
	roots := ft.eventKnowsRoots[parent]
	return &roots
}

func (ft *FlagTable) SetKnownRoots(e *Event, roots EventHashes) {
	ft.eventKnowsRoots[e.Hash()] = roots
	// TODO: is it last anyway?
	//ft.lastNodeEvent[e.Creator] = e.Hash()
	// TODO: save in store
}

/*
 * Poset's methods:
 */

// checkIfRoot is not safe for concurrent use.
func (p *Poset) checkIfRoot(e *Event) bool {
	var knownRoots EventHashes

	for parent := range e.Parents.All() {
		for root := range p.flagTable.KnownRootsOf(e, parent).All() {
			knownRoots.Add(root)
		}
	}
	p.flagTable.SetKnownRoots(e, knownRoots)

	stake := p.newStakeCounter()
	for h := range knownRoots.All() {
		parent := p.store.GetEvent(h)
		stake.Count(parent.Creator)
	}

	if !stake.HasMajority() {
		return false
	}

	frame := p.frame(p.state.LastFinishedFrameN + 1)
	frame.SetRoot(e.Hash())

	return true
}
