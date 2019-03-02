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

	nodes := newParentsValidator(e)
	ltime := newLamportTimeValidator(e)

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
	var frame *Frame
	if frame = p.checkIfRoot(e); frame == nil {
		return
	}
	p.setClothoCandidates(e, frame)

	// process matured frames where ClothoCandidates have become Clothos
	for n := p.state.LastFinishedFrameN + 1; n+3 <= frame.Index; n++ {
		if p.hasAtropos(n, frame.Index) {
			p.makeBlock(n)
		}
	}
}

// checkIfRoot checks root-conditions for new event
// and returns frame where event is root.
// It is not safe for concurrent use.
func (p *Poset) checkIfRoot(e *Event) *Frame {
	//log.Debugf("----- %s", e)
	knownRoots := eventsByFrame{}
	minFrame := p.state.LastFinishedFrameN
	for parent := range e.Parents {
		if !parent.IsZero() {
			frame, isRoot := p.FrameOfEvent(parent)
			if prev := p.store.GetEvent(parent); prev.Creator == e.Creator {
				minFrame = frame.Index
			}
			roots := frame.GetRootsOf(parent)
			knownRoots.Add(frame.Index, roots)
			if !isRoot || frame.Index <= minFrame {
				continue
			}
			frame = p.frame(frame.Index-1, false)
			roots = frame.GetRootsOf(parent)
			knownRoots.Add(frame.Index, roots)
		} else {
			roots := rootZero(e.Creator)
			knownRoots.Add(p.state.LastFinishedFrameN+1, roots)
		}
	}

	for _, fnum := range knownRoots.FrameNumsDesc() {
		if fnum < minFrame {
			continue
		}
		roots := knownRoots[fnum]
		frame := p.frame(fnum, true)
		frame.AddRootsOf(e.Hash(), roots)
		//log.Debugf(" %s knows %s at frame %d", e.Hash().String(), roots.String(), frame.Index)
		if p.hasMajority(roots) {
			frame = p.frame(fnum+1, true)
			frame.AddRootsOf(e.Hash(), rootFrom(e))
			//log.Debugf(" %s is root of frame %d", e.Hash().String(), frame.Index)
			return frame
		}
	}
	return nil
}

// setClothoCandidates checks clotho-conditions for seen by new root.
// It is not safe for concurrent use.
func (p *Poset) setClothoCandidates(root *Event, frame *Frame) {
	// check Clotho Candidates in previous frame
	prev := p.frame(frame.Index-1, false)
	// events from previous frame, reachable by root
	for seen, seenCreator := range prev.FlagTable[root.Hash()].Each() {
		// seen is CC already
		if prev.ClothoCandidates.Contains(seenCreator, seen) {
			continue
		}
		// seen is not root
		if !prev.FlagTable.IsRoot(seen) {
			continue
		}
		// all roots from frame, reach the seen
		roots := eventsByNode{}
		for root, creator := range frame.FlagTable.Roots().Each() {
			if prev.FlagTable.EventKnows(root, seenCreator, seen) {
				roots.AddOne(root, creator)
			}
		}
		// check CC-condition
		if p.hasTrust(roots) {
			prev.AddClothoCandidate(seen, seenCreator)
			//log.Debugf("CC: %s from %s", seen.String(), seenCreator.String())
		}
	}
}

// hasAtropos checks frame Clothos for Atropos condition.
// It is not safe for concurrent use.
// Algorithm 5 "Atropos Consensus Time Selection" implementation.
func (p *Poset) hasAtropos(frameNum, lastNum uint64) bool {
	var has = false
	frame := p.frame(frameNum, false)
CLOTHO:
	for clotho, clothoCreator := range frame.ClothoCandidates.Each() {
		if _, already := frame.Atroposes[clotho]; already {
			has = true
			continue CLOTHO
		}

		timeOfC := timestampsByEvent{}

		// for later frames
		for diff := uint64(1); diff <= (lastNum - frameNum); diff++ {
			later := p.frame(frameNum+diff, false)
			for r := range later.FlagTable.Roots().Each() {
				if diff == 1 {
					if frame.FlagTable.EventKnows(r, clothoCreator, clotho) {
						timeOfC[r] = p.store.GetEvent(r).LamportTime
					}
				} else {
					// the set of root in prev frame that r can be happened-before with 2n/3 condition
					prev := p.frame(later.Index-1, false)
					S := prev.GetRootsOf(r)
					// time reselection (algorithm 6 "Consensus Time Reselection" implementation)
					counter := timeCounter{}
					for r := range S.Each() {
						if t, ok := timeOfC[r]; ok {
							counter.Add(t)
						}
					}
					T := counter.MaxMin()
					// votes for reselected time
					K := eventsByNode{}
					for r, n := range S.Each() {
						if t, ok := timeOfC[r]; ok && t == T {
							K.AddOne(r, n)
						}
					}

					if diff%3 > 0 && p.hasMajority(K) {
						frame.SetAtropos(clotho, T)
						has = true
						continue CLOTHO
					}

					timeOfC[r] = T
				}
			}
		}
	}
	return has
}

func (p *Poset) makeBlock(frameNum uint64) {
	/*
		frame := p.frame(frameNum, false)

		// NOTE: actually not every Clotho is Atropos, fix it
		var orderedAtroposes Events
		for clotho := range frame.ClothoCandidates.Each() {
			sortedAtropos = append(atroposes, p.store.GetEvent(clotho))
		}
		sort.Sort(orderedAtroposes)

		for _, a := range orderedAtroposes {

		}
	*/
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
