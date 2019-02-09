package posposet

import (
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// Poset processes events to get consensus.
type Poset struct {
	store     *Store
	state     *State
	flagTable *FlagTable
	frames    map[uint64]*Frame

	processingWg   sync.WaitGroup
	processingDone chan struct{}

	newEventsCh      chan *Event
	incompleteEvents map[EventHash]*Event
}

// New creates Poset instance.
func New(store *Store) *Poset {
	const buffSize = 10

	p := &Poset{
		store: store,

		newEventsCh:      make(chan *Event, buffSize),
		incompleteEvents: make(map[EventHash]*Event),
	}

	p.bootstrap()
	return p
}

// Start starts events processing. It is not safe for concurrent use.
func (p *Poset) Start() {
	if p.processingDone != nil {
		return
	}
	p.processingDone = make(chan struct{})
	p.processingWg.Add(1)
	go func() {
		defer p.processingWg.Done()
		log.Debug("Start of events processing ...")
		for {
			select {
			case <-p.processingDone:
				log.Debug("Stop of events processing ...")
				return
			case e := <-p.newEventsCh:
				p.onNewEvent(e)
			}
		}
	}()
}

// Stop stops events processing. It is not safe for concurrent use.
func (p *Poset) Stop() {
	if p.processingDone == nil {
		return
	}
	close(p.processingDone)
	p.processingWg.Wait()
	p.processingDone = nil
}

// PushEvent takes event into processing. Event order doesn't matter.
func (p *Poset) PushEvent(e Event) {
	initEventIdx(&e)

	p.newEventsCh <- &e
}

// onNewEvent runs consensus calc from new event. It is not safe for concurrent use.
func (p *Poset) onNewEvent(e *Event) {
	if p.store.HasEvent(e.Hash()) {
		log.WithField("event", e).Warnf("Event had received already")
		return
	}

	// unique parent nodes checker
	parentNodes := make(map[common.Address]struct{})
	isParentNodeUnique := func(e *Event, node common.Address) bool {
		if _, ok := parentNodes[node]; ok {
			log.Warnf("Event %s has double refer to node %s, so rejected", e.Hash().ShortString(), node.String())
			return false
		}
		parentNodes[node] = struct{}{}
		return true
	}

	// fill event's parents index or hold it as incompleted
	for hash := range e.Parents.All() {
		if hash.IsZero() {
			// first event of node
			if !isParentNodeUnique(e, e.Creator) {
				return
			}
			continue
		}
		parent := e.parents[hash]
		if parent == nil {
			parent = p.store.GetEvent(hash)
			if parent == nil {
				//log.WithField("event", e).Debug("Event's parent has not received yet")
				p.incompleteEvents[e.Hash()] = e
				return
			}
			e.parents[hash] = parent
		}
		if !isParentNodeUnique(e, parent.Creator) {
			return
		}
	}

	// parents OK
	p.store.SetEvent(e)
	p.consensus(e)

	// now child events may become complete, check it again
	for hash, child := range p.incompleteEvents {
		if parent, ok := child.parents[e.Hash()]; ok && parent == nil {
			child.parents[e.Hash()] = e
			delete(p.incompleteEvents, hash)
			p.onNewEvent(child)
		}
	}
}

// consensus is not safe for concurrent use.
func (p *Poset) consensus(e *Event) {
	if !p.checkIfRoot(e) {
		return
	}
	if !p.checkIfClotho(e) {
		return
	}
}

func (p *Poset) checkIfClotho(e *Event) bool {
	// TODO: implement it
	return false
}

/*
 * Utils:
 */

// initEventIdx initializes internal index of parents.
func initEventIdx(e *Event) {
	if e.parents != nil {
		return
	}
	e.parents = make(map[EventHash]*Event, e.Parents.Len())
	for hash := range e.Parents.All() {
		e.parents[hash] = nil
	}
}
