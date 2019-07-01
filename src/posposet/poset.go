package posposet

import (
	"math"
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
	frames map[uint64]*Frame

	processingWg   sync.WaitGroup
	processingDone chan struct{}

	newEventsCh chan hash.Event
	onNewEvent  func(*inter.Event) // onNewEvent runs consensus calc from new event

	newBlockCh   chan uint64
	onNewBlockMu sync.RWMutex
	onNewBlock   func(blockNumber uint64)

	firstDescendantsSeq []int64
	lastAncestorsSeq    []int64
	recentEvents        []hash.Event      // index is a member index.
	membersByPears      map[hash.Peer]int // mapping creator id -> member num
	members             []*Member

	logger.Instance
}

// New creates Poset instance.
// It does not start any process.
func New(store *Store, input EventSource, membersNumber int) *Poset {
	const buffSize = 10

	firstDescendantsSeq := make([]int64, membersNumber)
	lastAncestorsSeq := make([]int64, membersNumber)
	recentEvents := make([]hash.Event, membersNumber)

	for i := 0; i < membersNumber; i++ {
		firstDescendantsSeq[i] = math.MaxInt32
		lastAncestorsSeq[i] = -1
	}

	p := &Poset{
		store:  store,
		input:  input,
		frames: make(map[uint64]*Frame),

		newEventsCh: make(chan hash.Event, buffSize),

		newBlockCh: make(chan uint64, buffSize),

		Instance: logger.MakeInstance(),

		firstDescendantsSeq: firstDescendantsSeq,
		lastAncestorsSeq:    lastAncestorsSeq,
		recentEvents:        recentEvents,
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
	lastFinished := p.state.LastFinishedFrameN()
	for n := p.state.LastFinishedFrameN() + 1; n+3 <= frame.Index; n++ {
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
	if p.state.LastFinishedFrameN() < lastFinished {
		p.state.LastFinishedFrame(lastFinished)
		p.saveState()
		p.Debugf("consensus: lastFinishedFrameN is %d", p.state.LastFinishedFrameN())
	}

	// clean old frames
	for i := range p.frames {
		if i+stateGap < p.state.LastFinishedFrameN() {
			delete(p.frames, i)
		}
	}
}

// checkIfRoot checks root-conditions for new event
// and returns frame where event is root.
// It is not safe for concurrent use.
func (p *Poset) checkIfRoot(e *Event) *Frame {
	knownRoots := eventsByFrame{}
	minFrame := p.state.LastFinishedFrameN() + 1
	for parent := range e.Parents {
		if !parent.IsZero() {
			frame, isRoot := p.FrameOfEvent(parent)
			if frame == nil || frame.Index <= p.state.LastFinishedFrameN() {
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
	all := inter.Events{}
	// foreach stale frame
	for n := start; n <= stop; n++ {
		frame := p.frames[n]
		// extract events
		for e := range frame.FlagTable {
			if !frame.FlagTable.IsRoot(e) {
				all = append(all, p.GetEvent(e).Event)
			}
		}
		// and replace stale frame with blank
		p.frames[n] = &Frame{
			Index:            n,
			FlagTable:        FlagTable{},
			ClothoCandidates: EventsByPeer{},
			Atroposes:        TimestampsByEvent{},
			Balances:         newBalance,
		}
	}
	// recalc consensus (without frame saving)
	events := all.ByParents()
	if len(events) != 0 {
		for i := range events {
			p.consensus(events[i])
		}
	}

	// save fresh frame
	for n := start; n <= stop; n++ {
		frame := p.frames[n]

		p.setFrameSaving(frame)
		frame.Save()
	}
}

func (p *Poset) findMemberNumber(creator hash.Peer) (int, bool) {
	num, ok := p.membersByPears[creator]
	if !ok {
		return 0, false
	}
	return num, true
}

func (p *Poset) getMember(id int) (*Member, bool) {
	if id >= len(p.members) {
		return nil, false
	}

	return p.members[id], true
}

func (p *Poset) highestEventFromMember(member int) (hash.Event, error) {
	if member >= len(p.recentEvents) {
		return hash.Event{}, ErrInvalidMemberNum
	}
	return p.recentEvents[member], nil
}

func (p *Poset) memberLastAncestorSeq(member int) int64 {
	if member >= len(p.lastAncestorsSeq) {
		return -1 // default value to last ancestor sequence.
	}
	return p.lastAncestorsSeq[member]
}

// sufficientCoherence calculates "sufficient coherence" between the events.
// The event1.lastAncestorsSeq array remembers the sequence number of the last
// event by each member that is an ancestor of event1. The array for
// event2.FirstDescendantsSeq remembers the sequence number of the earliest
// event by each member that is a descendant of event2. Compare the two arrays,
// and find how many elements in the event1.lastAncestorsSeq array are greater
// than or equal to the corresponding element of the event2.FirstDescendantsSeq
// array. If there are more than 2n/3 such matches, then the event1 and event2
// have achieved sufficient coherency.
func (p *Poset) sufficientCoherence(event1, event2 *Event) bool {
	if len(event1.LastAncestorsSeq) != len(event2.FirstDescendantsSeq) {
		return false
	}

	counter := 0
	for k := range event1.LastAncestorsSeq {
		if event2.FirstDescendantsSeq[k] <= event1.LastAncestorsSeq[k] {
			counter++
		}
	}

	if counter >= len(event1.LastAncestorsSeq)*2/3 {
		return true
	}

	return false
}

func (p *Poset) fillEventSequences(event *Event) {
	memberNumber, ok := p.findMemberNumber(event.Creator)
	if !ok {
		return
	}

	var (
		foundSelfParent  bool
		foundOtherParent bool
	)

	getOtherParent := func() *Event {
		// TODO: we need to determine the number of other parents in the future.
		op := event.OtherParents()[0] // take a first other parent.
		return p.GetEvent(op)
	}

	initLastAncestors := func() {
		if len(event.LastAncestors) == len(p.members) {
			return
		}
		event.LastAncestors = make([]hash.Event, len(p.members))
	}

	selfParent, found := event.SelfParent()
	if found {
		foundSelfParent = true
	}

	otherParents := event.OtherParents()
	if otherParents.Len() > 0 {
		foundOtherParent = true
	}

	if !foundSelfParent && !foundOtherParent {
		event.LastAncestorsSeq = p.lastAncestorsSeq

		highestEvent, err := p.highestEventFromMember(memberNumber)
		if err != nil {
			p.Fatal(err.Error())
			return
		}

		initLastAncestors()
		event.LastAncestors[memberNumber] = highestEvent
	} else if !foundSelfParent {
		parent := getOtherParent()
		event.LastAncestors = parent.LastAncestors
		event.LastAncestorsSeq = parent.LastAncestorsSeq
	} else if !foundOtherParent {
		parent := p.GetEvent(selfParent)
		event.LastAncestors = parent.LastAncestors
		event.LastAncestorsSeq = parent.LastAncestorsSeq
	} else {
		sp := p.GetEvent(selfParent)
		event.LastAncestors = sp.LastAncestors
		event.LastAncestorsSeq = sp.LastAncestorsSeq

		otherParent := getOtherParent()

		for i := 0; i < len(p.members); i++ {
			if event.LastAncestorsSeq[i] >= otherParent.LastAncestorsSeq[i] {
				event.LastAncestors[i] = otherParent.LastAncestors[i]
				event.LastAncestorsSeq[i] = otherParent.LastAncestorsSeq[i]
			}
		}
	}

	event.FirstDescendantsSeq = p.firstDescendantsSeq
	event.FirstDescendants = p.recentEvents

	event.LastAncestors[memberNumber] = event.Hash()
	event.FirstDescendants[memberNumber] = event.Hash()

	event.FirstDescendantsSeq[memberNumber] =
		p.memberLastAncestorSeq(memberNumber)
	event.LastAncestorsSeq[memberNumber] =
		p.memberLastAncestorSeq(memberNumber)
}
