package posposet

import (
	"sort"
	"sync"

	"github.com/pkg/errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

// Poset processes events to get consensus.
type Poset struct {
	store  *Store
	state  *State
	input  EventSource
	frames *sync.Map

	processingWg   sync.WaitGroup
	processingDone chan struct{}

	newEventsCh chan hash.Event
	onNewEvent  func(*inter.Event) // onNewEvent runs consensus calc from new event

	newBlockCh   chan uint64
	onNewBlockMu sync.RWMutex
	onNewBlock   func(blockNumber uint64)

	logger.Instance
}

// New creates Poset instance.
// It does not start any process.
func New(store *Store, input EventSource) *Poset {
	const buffSize = 10

	p := &Poset{
		store:  store,
		input:  input,
		frames: new(sync.Map),

		newEventsCh: make(chan hash.Event, buffSize),

		newBlockCh: make(chan uint64, buffSize),

		Instance: logger.MakeInstance(),
	}

	// event order matter: parents first
	p.onNewEvent = func(e *inter.Event) {
		if e == nil {
			panic("got unsaved event")
		}
		p.consensus(e)
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
		// log.Debug("Start of events processing ...")
		for {
			select {
			case <-p.processingDone:
				// log.Debug("Stop of events processing ...")
				return
			case e := <-p.newEventsCh:
				event := p.input.GetEvent(e)
				p.onNewEvent(event)

			case num := <-p.newBlockCh:
				p.onNewBlockMu.RLock()
				if p.onNewBlock != nil {
					p.onNewBlock(num)
				}
				p.onNewBlockMu.RUnlock()
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

// PushEvent takes event into processing.
// Event order matter: parents first.
func (p *Poset) PushEvent(e hash.Event) {
	p.newEventsCh <- e
}

// OnNewBlock sets (or replaces if override) a callback that is called on new block.
// Returns an error if can not.
func (p *Poset) OnNewBlock(callback func(blockNumber uint64), override bool) error {
	// TODO: support multiple subscribers later
	p.onNewBlockMu.Lock()
	defer p.onNewBlockMu.Unlock()
	if !override && p.onNewBlock != nil {
		return errors.New("callback already registered")
	}

	p.onNewBlock = callback
	return nil
}

// consensus is not safe for concurrent use.
func (p *Poset) consensus(event *inter.Event) {
	p.Debugf("consensus: start %s", event.String())
	e := &Event{
		Event: event,
	}

	var frame *Frame
	if frame = p.checkIfRoot(e); frame == nil {
		return
	}
	p.Debugf("consensus: %s is root", event.String())

	p.setClothoCandidates(e, frame)

	// process matured frames where ClothoCandidates have become Clothos
	var ordered inter.Events
	lastFinished := p.state.LastFinishedFrameN
	for n := p.state.LastFinishedFrameN + 1; n+3 <= frame.Index; n++ {
		if p.hasAtropos(n, frame.Index) {
			p.Debugf("consensus: make new block %d from frame %d", p.state.LastBlockN+1, n)
			events := p.topologicalOrdered(n)
			block := inter.NewBlock(p.state.LastBlockN+1, events)
			p.store.SetEventsBlockNum(block.Index, events...)
			p.store.SetBlock(block)
			p.state.LastBlockN = block.Index
			p.saveState()
			if p.newBlockCh != nil {
				p.newBlockCh <- p.state.LastBlockN
			}

			// TODO: fix it
			lastFinished = n // NOTE: are every event of prev frame there in block? (No)

			ordered = append(ordered, events...)
		}
	}

	// balances changes
	applyAt := p.frame(frame.Index+stateGap, true)
	state := p.store.StateDB(applyAt.Balances)
	p.applyTransactions(state, ordered)
	p.applyRewards(state, ordered)
	balances, err := state.Commit(true)
	if err != nil {
		p.Fatal(err)
	}
	if applyAt.SetBalances(balances) {
		p.Debugf("consensus: new state [%d]%s --> [%d]%s", frame.Index, frame.Balances.String(), applyAt.Index, balances.String())
		p.reconsensusFromFrame(applyAt.Index, balances)
	}

	// save finished frames
	if p.state.LastFinishedFrameN < lastFinished {
		p.state.LastFinishedFrameN = lastFinished
		p.saveState()
		p.Debugf("consensus: lastFinishedFrameN is %d", p.state.LastFinishedFrameN)
	}

	// clean old frames
	p.frames.Range(func(key, value interface{}) bool {
		i, ok := key.(uint64)
		if !ok {
			p.Fatal(ErrIncorrectFrameKeyType)
		}

		if i+stateGap < p.state.LastFinishedFrameN {
			p.frames.Delete(i)
		}

		return true
	})
}

// checkIfRoot checks root-conditions for new event
// and returns frame where event is root.
// It is not safe for concurrent use.
func (p *Poset) checkIfRoot(e *Event) *Frame {
	knownRoots := eventsByFrame{}
	minFrame := p.state.LastFinishedFrameN + 1
	for parent := range e.Parents {
		if !parent.IsZero() {
			frame, isRoot := p.FrameOfEvent(parent)
			if frame == nil || frame.Index <= p.state.LastFinishedFrameN {
				p.Warnf("Parent %s of %s is too old. Skipped", parent.String(), e.String())
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

	var (
		frame  *Frame
		isRoot bool
	)
	for _, fnum := range knownRoots.FrameNumsDesc() {
		if fnum < minFrame {
			break
		}
		roots := knownRoots[fnum]
		frame = p.frame(fnum, true)
		frame.AddRootsOf(e.Hash(), roots)
		// log.Debugf(" %s knows %s at frame %d", e.Hash().String(), roots.String(), frame.Index)
		if isRoot = p.hasMajority(frame, roots); isRoot {
			frame = p.frame(fnum+1, true)
			// log.Debugf(" %s is root of frame %d", e.Hash().String(), frame.Index)
			break
		}
	}
	if !p.isEventValid(e, frame) {
		return nil
	}
	p.store.SetEventFrame(e.Hash(), frame.Index)
	if isRoot {
		frame.AddRootsOf(e.Hash(), rootFrom(e))
		return frame
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
			// log.Debugf("CC: %s from %s", seen.String(), seenCreator.String())
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
						// log.Debugf("ATROPOS %s of frame %d", clotho.String(), frame.Index)
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
func (p *Poset) topologicalOrdered(frameNum uint64) (chain inter.Events) {
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
		for _, e := range ee {
			chain = append(chain, e.Event)
		}
		chain = append(chain, atropos.Event)
	}

	return
}

// collectParents recursive collects Events of Atropos.
func (p *Poset) collectParents(a *Event, res *Events, already hash.Events) {
	for hash_ := range a.Parents {
		if hash_.IsZero() {
			continue
		}
		if already.Contains(hash_) {
			continue
		}
		f, _ := p.FrameOfEvent(hash_)
		if _, ok := f.Atroposes[hash_]; ok {
			continue
		}

		e := p.GetEvent(hash_)
		e.consensusTime = a.consensusTime
		*res = append(*res, e)
		already.Add(hash_)
		p.collectParents(e, res, already)
	}
}

// reconsensusFromFrame recalcs consensus of frames.
func (p *Poset) reconsensusFromFrame(start uint64, newBalance hash.Hash) {
	stop := p.frameNumLast()
	var all inter.Events
	// foreach stale frame
	for n := start; n <= stop; n++ {
		frame := p.mustFrameLoad(n)
		// extract events
		for e := range frame.FlagTable {
			if !frame.FlagTable.IsRoot(e) {
				all = append(all, p.GetEvent(e).Event)
			}
		}
		// and replace stale frame with blank
		p.frames.Store(n, &Frame{
			Index:            n,
			FlagTable:        FlagTable{},
			ClothoCandidates: EventsByPeer{},
			Atroposes:        TimestampsByEvent{},
			Balances:         newBalance,
		})
	}
	// recalc consensus (without frame saving)
	for _, e := range all.ByParents() {
		p.consensus(e)
	}
	// save fresh frame
	for n := start; n <= stop; n++ {
		frame := p.mustFrameLoad(n)

		p.setFrameSaving(frame)
		frame.Save()
	}
}
