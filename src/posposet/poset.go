package posposet

import (
	"sync"
)

// Poset processes events to get consensus.
type Poset struct {
	store     *Store
	state     *State
	flagTable FlagTable
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
		store:  store,
		frames: make(map[uint64]*Frame),

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
		//log.Debug("Start of events processing ...")
		for {
			select {
			case <-p.processingDone:
				//log.Debug("Stop of events processing ...")
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

	nodes := newParentNodesInspector(e)
	ltime := newParentLamportTimeInspector(e)

	// fill event's parents index or hold it as incompleted
	for hash := range e.Parents {
		if hash.IsZero() {
			// first event of node
			if !nodes.IsParentUnique(e.Creator) {
				return
			}
			if !ltime.IsGreaterThan(0) {
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
		if !nodes.IsParentUnique(parent.Creator) {
			return
		}
		if !ltime.IsGreaterThan(parent.LamportTime) {
			return
		}
	}
	if !nodes.HasSelfParent() {
		return
	}
	if !ltime.IsSequential() {
		return
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

// checkIfRoot checks root-conditions for new event.
// Event.parents should be filled.
// It is not safe for concurrent use.
func (p *Poset) checkIfRoot(e *Event) bool {
	//log.Debugf("----- %s", e)

	// known roots groupped by frame num
	knownRootsDirect := Roots{}
	knownRootsThrough := Roots{}
	for parent := range e.Parents {
		if !parent.IsZero() {
			frame, isRoot := p.FrameOfEvent(parent)
			if frame == nil {
				continue
			}
			roots := frame.EventRootsGet(parent)
			knownRootsDirect.Add(frame.Index, roots)
			if !isRoot {
				continue
			}
			frame = p.frame(frame.Index-1, false)
			if frame == nil {
				continue
			}
			roots = frame.EventRootsGet(parent)
			knownRootsThrough.Add(frame.Index, roots)
		} else {
			roots := rootZero(e.Creator)
			knownRootsDirect.Add(p.state.LastFinishedFrameN+1, roots)
		}
	}
	// use parent's frames only
	knownRoots := Roots{}
	for n, _ := range knownRootsDirect {
		knownRoots.Add(n, knownRootsDirect[n])
		knownRoots.Add(n, knownRootsThrough[n])
	}

	for _, fnum := range knownRoots.FrameNumsDesc() {
		roots := knownRoots[fnum]

		frame := p.frame(fnum, true)
		frame.EventRootsAdd(e.Hash(), roots)
		//log.Debugf(" %s knows %s at frame %d", e.Hash().String(), roots.String(), frame.Index)

		if p.hasMajority(roots) {
			frame = p.frame(fnum+1, true)
			frame.EventRootsAdd(e.Hash(), rootFrom(e))
			//log.Debugf(" %s is root of frame %d", e.Hash().String(), frame.Index)
			return true
		}
	}
	return false
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
	e.parents = make(map[EventHash]*Event, len(e.Parents))
	for hash := range e.Parents {
		e.parents[hash] = nil
	}
}
