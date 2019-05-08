package posposet

import (
	"sort"
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// Poset processes events to get consensus.
type Poset struct {
	store  *Store
	state  *State
	input  EventSource
	frames map[uint64]*Frame

	processingWg   sync.WaitGroup
	processingDone chan struct{}

	newEventsCh      chan hash.Event
	incompleteEvents map[hash.Event]*Event

	NewBlockCh chan uint64
}

// New creates Poset instance.
// It does not start any process.
func New(store *Store, input EventSource) *Poset {
	const buffSize = 10

	p := &Poset{
		store:  store,
		input:  input,
		frames: make(map[uint64]*Frame),

		newEventsCh:      make(chan hash.Event, buffSize),
		incompleteEvents: make(map[hash.Event]*Event),
	}

	return p
}

// Start starts events processing.
func (p *Poset) Start() {
	if p.processingDone != nil {
		return
	}

	p.Bootstrap()

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
				event := p.GetEvent(e)
				p.onNewEvent(event)
			}
		}
	}()
}

// Stop stops events processing.
func (p *Poset) Stop() {
	if p.processingDone == nil {
		return
	}
	close(p.processingDone)
	p.processingWg.Wait()
	p.processingDone = nil
}

// PushEvent takes event into processing. Event order doesn't matter.
func (p *Poset) PushEvent(e hash.Event) {
	p.newEventsCh <- e
}

// onNewEvent runs consensus calc from new event. It is not safe for concurrent use.
func (p *Poset) onNewEvent(e *Event) {
	if p.store.GetEventFrame(e.Hash()) != nil {
		log.WithField("event", e).Warnf("Event had received already")
		return
	}

	if e.parents == nil {
		e.parents = make(map[hash.Event]*Event, len(e.Parents))
		for hash := range e.Parents {
			e.parents[hash] = nil
		}
	}

	// TODO: move validation to level up for sync error
	nodes := newParentsValidator(e)
	ltime := newLamportTimeValidator(e)

	// fill event's parents index or hold it as incompleted
	for pHash := range e.Parents {
		if pHash.IsZero() {
			// first event of node
			if !nodes.IsParentUnique(e.Creator) {
				return
			}
			if !ltime.IsGreaterThan(0) {
				return
			}
			continue
		}
		parent := e.parents[pHash]
		if parent == nil {
			if p.store.GetEventFrame(pHash) == nil {
				//log.WithField("event", e).Debug("Event's parent has not received yet")
				p.incompleteEvents[e.Hash()] = e
				return
			}
			parent = p.GetEvent(pHash)
			e.parents[pHash] = parent
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
	const X = 3 // TODO: remove this magic number

	var frame *Frame
	if frame = p.checkIfRoot(e); frame == nil {
		return
	}
	p.setClothoCandidates(e, frame)

	// process matured frames where ClothoCandidates have become Clothos
	var ordered Events
	lastFinished := p.state.LastFinishedFrameN
	for n := p.state.LastFinishedFrameN + 1; n+3 <= frame.Index; n++ {
		if p.hasAtropos(n, frame.Index) {
			events := p.topologicalOrdered(n)
			p.state.LastBlockN = p.makeBlock(events)
			p.saveState()

			if p.NewBlockCh != nil {
				p.NewBlockCh <- p.state.LastBlockN
			}

			// TODO: fix it
			lastFinished = n // NOTE: are every event of prev frame there in block? (No)

			ordered = append(ordered, events...)
		}
	}

	// apply transactions
	next := p.frame(frame.Index+X, true)
	if next.SetBalances(p.applyTransactions(frame.Balances, ordered)) {
		p.reconsensusFromFrame(next.Index)
	}

	// save finished frames
	if p.state.LastFinishedFrameN < lastFinished {
		p.state.LastFinishedFrameN = lastFinished
		p.saveState()
	}

	// clean old frames
	for i := range p.frames {
		if i+X < p.state.LastFinishedFrameN {
			delete(p.frames, i)
		}
	}
}

// checkIfRoot checks root-conditions for new event
// and returns frame where event is root.
// It is not safe for concurrent use.
func (p *Poset) checkIfRoot(e *Event) *Frame {
	//log.Debugf("----- %s", e)
	knownRoots := eventsByFrame{}
	minFrame := p.state.LastFinishedFrameN + 1
	for parent := range e.Parents {
		if !parent.IsZero() {
			frame, isRoot := p.FrameOfEvent(parent)
			if frame == nil || frame.Index <= p.state.LastFinishedFrameN {
				log.Warnf("Parent %s of %s is too old. Skipped", parent.String(), e.String())
				// NOTE: is it possible some participants got this event before parent outdated?
				continue
			}
			if prev := p.GetEvent(parent); prev.Creator == e.Creator {
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
			knownRoots.Add(minFrame, roots)
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
		f := p.frame(fnum, false)
		if p.hasMajority(f, roots) {
			frame = p.frame(fnum+1, true)
			frame.AddRootsOf(e.Hash(), rootFrom(e))
			p.store.SetEventFrame(e.Hash(), frame.Index)
			//log.Debugf(" %s is root of frame %d", e.Hash().String(), frame.Index)
			return frame
		}
		p.store.SetEventFrame(e.Hash(), frame.Index)
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
		roots := EventsByPeer{}
		for root, creator := range frame.FlagTable.Roots().Each() {
			if prev.FlagTable.EventKnows(root, seenCreator, seen) {
				roots.AddOne(root, creator)
			}
		}
		// check CC-condition
		if p.hasTrust(frame, roots) {
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

		candidateTime := TimestampsByEvent{}

		// for later frames
		for diff := uint64(1); diff <= (lastNum - frameNum); diff++ {
			later := p.frame(frameNum+diff, false)
			for r := range later.FlagTable.Roots().Each() {
				if diff == 1 {
					if frame.FlagTable.EventKnows(r, clothoCreator, clotho) {
						candidateTime[r] = p.GetEvent(r).LamportTime
					}
				} else {
					// the set of root in prev frame that r can be happened-before with 2n/3 condition
					prev := p.frame(later.Index-1, false)
					S := prev.GetRootsOf(r)
					// time reselection (algorithm 6 "Consensus Time Reselection" implementation)
					counter := timeCounter{}
					for r := range S.Each() {
						if t, ok := candidateTime[r]; ok {
							counter.Add(t)
						}
					}
					T := counter.MaxMin()
					// votes for reselected time
					K := EventsByPeer{}
					for r, n := range S.Each() {
						if t, ok := candidateTime[r]; ok && t == T {
							K.AddOne(r, n)
						}
					}

					if diff%3 > 0 && p.hasMajority(prev, K) {
						//log.Debugf("ATROPOS %s of frame %d", clotho.String(), frame.Index)
						frame.SetAtropos(clotho, T)
						has = true
						continue CLOTHO
					}

					candidateTime[r] = T
				}
			}
		}
	}
	return has
}

// topologicalOrdered sorts events to chain with Atropos consensus time.
func (p *Poset) topologicalOrdered(frameNum uint64) (chain Events) {
	frame := p.frame(frameNum, false)

	var atroposes Events
	for atropos, consensusTime := range frame.Atroposes {
		e := p.GetEvent(atropos)
		e.consensusTime = consensusTime
		atroposes = append(atroposes, e)
	}
	sort.Sort(atroposes)

	already := hash.Events{}
	for _, atropos := range atroposes {
		ee := Events{}
		p.collectParents(atropos, &ee, already)
		sort.Sort(ee)
		chain = append(chain, ee...)
		chain = append(chain, atropos)
	}

	return
}

// collectParents recursive collects Events of Atropos.
func (p *Poset) collectParents(a *Event, res *Events, already hash.Events) {
	for hash := range a.Parents {
		if hash.IsZero() {
			continue
		}
		if already.Contains(hash) {
			continue
		}
		f, _ := p.FrameOfEvent(hash)
		if _, ok := f.Atroposes[hash]; ok {
			continue
		}

		e := p.GetEvent(hash)
		e.consensusTime = a.consensusTime
		*res = append(*res, e)
		already.Add(hash)
		p.collectParents(e, res, already)
	}
}

// makeBlock makes main chain block from topological ordered events.
func (p *Poset) makeBlock(ordered Events) uint64 {
	events := make(hash.EventsSlice, len(ordered))
	for i, e := range ordered {
		events[i] = e.Hash()
	}

	b := &Block{
		Index:  p.state.LastBlockN + 1,
		Events: events,
	}
	p.store.SetBlock(b)
	// TODO: notify external systems (through chan)
	return b.Index
}

// reconsensusFromFrame recalcs consensus of frames.
// It is not safe for concurrent use.
func (p *Poset) reconsensusFromFrame(start uint64) {
	stop := p.frameNumLast()
	var all inter.Events
	// foreach stale frame
	for n := start; n <= stop; n++ {
		frame := p.frames[n]
		// extract events
		for e := range frame.FlagTable {
			if !frame.FlagTable.IsRoot(e) {
				all = append(all, &p.GetEvent(e).Event)
			}
		}
		// and replace stale frame with blank
		p.frames[n] = &Frame{
			Index:            n,
			FlagTable:        FlagTable{},
			ClothoCandidates: EventsByPeer{},
			Atroposes:        TimestampsByEvent{},
			Balances:         frame.Balances,
		}
	}
	// recalc consensus
	for _, e := range all.ByParents() {
		p.consensus(&Event{
			Event: *e,
		})
	}
	// foreach fresh frame
	for n := start; n <= stop; n++ {
		frame := p.frames[n]
		// save fresh frame
		p.setFrameSaving(frame)
		frame.Save()
	}
}
