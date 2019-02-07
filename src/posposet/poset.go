package posposet

import (
	"fmt"
	"sync"
)

// Poset processes events to get consensus.
type Poset struct {
	store *Store
	state *State

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
		log.Debug("Start processing ...")
		for {
			select {
			case <-p.processingDone:
				log.Debug("Stop processing ...")
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
func (p *Poset) PushEvent(e Event) error {
	err := initEventIdx(&e)
	if err != nil {
		return err
	}

	p.newEventsCh <- &e
	return nil
}

// onNewEvent runs consensus calc from new event. It is not safe for concurrent use.
func (p *Poset) onNewEvent(e *Event) {
	if p.store.HasEvent(e.Hash()) {
		log.WithField("event", e).Warnf("Event had received already")
		return
	}

	// fill event's parents index or hold it as incompleted
	for i, hash := range e.Parents {
		if i == 0 && hash.IsZero() {
			// first event from address
			continue
		}
		if parent := e.parents[hash]; parent == nil {
			parent = p.store.GetEvent(hash)
			if parent != nil {
				e.parents[hash] = parent
			} else {
				//log.WithField("event", e).Warn("Event's parent had not received yet")
				p.incompleteEvents[e.Hash()] = e
				return
			}
		}
	}
	p.store.SetEvent(e)
	p.consensus(e)

	// check child events complete
	for hash, incompleted := range p.incompleteEvents {
		if parent, ok := incompleted.parents[e.Hash()]; ok && parent == nil {
			delete(p.incompleteEvents, hash)
			p.onNewEvent(incompleted)
		}
	}
}

// consensus is not safe for concurrent use.
func (p *Poset) consensus(e *Event) {
	if p.checkIfRoot(e) {
		log.WithField("event", e).Debug("IS ROOT!")
	}
}

// checkIfRoot is not safe for concurrent use.
func (p *Poset) checkIfRoot(e *Event) bool {
	log.WithField("event", e).Debug("HERE")
	frame := p.frame(p.state.CurrentFrameN - 1)
	stakes := p.newStakes(frame)

	forEachParents := func(e *Event) {
		for hash, parent := range e.parents {
			if hash.IsZero() { // first event of node
				stakes.Count(e.Creator)
				continue
			}
			// read parent from store
			if parent == nil {
				parent = p.store.GetEvent(hash)
				if err := initEventIdx(parent); err != nil {
					panic(err)
				}
			}
			if frame.IsRoot(parent.Hash()) {
				stakes.Count(parent.Creator)
			}
		}
	}

	forEachParents(e)

	return stakes.HasMajority()
}

/*
 * Utils:
 */

func initEventIdx(e *Event) error {
	if e == nil {
		return fmt.Errorf("Event not found")
	}
	// internal parents index initialization
	e.parents = make(map[EventHash]*Event, len(e.Parents))
	for _, hash := range e.Parents {
		if _, ok := e.parents[hash]; ok {
			return fmt.Errorf("Event has double parents: %s", hash.String())
		}
		e.parents[hash] = nil
	}
	return nil
}
