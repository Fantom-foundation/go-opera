package posposet

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

// Poset processes events to get consensus.
type Poset struct {
	store *Store
	input EventSource
	state *State
	superFrame

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
		store: store,
		input: input,

		superFrame: *newSuperFrame(),

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

	// TODO: fill structs for strongly-see

	// TODO: try election to decide

	// TODO: order events to block if decided

	/* OLD :
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
	*/
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

// reconsensusFromFrame recalcs consensus of frames.
func (p *Poset) reconsensusFromFrame(start uint64, newBalance hash.Hash) {
	stop := p.frameNumLast()
	var all inter.Events
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
	for _, e := range all.ByParents() {
		p.consensus(e)
	}
	// save fresh frame
	for n := start; n <= stop; n++ {
		frame := p.frames[n]

		p.setFrameSaving(frame)
		frame.Save()
	}
}
