package posposet

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/election"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/vector"
)

// Poset processes events to get consensus.
type Poset struct {
	store *Store
	input EventSource
	*checkpoint
	superFrame

	election *election.Election
	events   *vector.Index

	processingWg   sync.WaitGroup
	processingDone chan struct{}

	newEventsCh chan hash.Event
	onNewEvent  func(*inter.Event) // onNewEvent runs consensus calc from new event

	newBlockCh   chan idx.Block
	onNewBlockMu sync.RWMutex
	onNewBlock   func(num idx.Block)

	logger.Instance
}

// New creates Poset instance.
// It does not start any process.
func New(store *Store, input EventSource) *Poset {
	const buffSize = 10

	p := &Poset{
		store: store,
		input: input,

		newEventsCh: make(chan hash.Event, buffSize),

		newBlockCh: make(chan idx.Block, buffSize),

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
func (p *Poset) OnNewBlock(callback func(blockNumber idx.Block), override bool) error {
	// TODO: support multiple subscribers later
	p.onNewBlockMu.Lock()
	defer p.onNewBlockMu.Unlock()
	if !override && p.onNewBlock != nil {
		return errors.New("callback already registered")
	}

	p.onNewBlock = callback
	return nil
}

// fills consensus-related fields: Frame, IsRoot, MedianTimestamp, GasLeft
// returns nil if event should be dropped
func (p *Poset) Prepare(e *inter.Event) *inter.Event {
	if e.Epoch != p.SuperFrameN {
		p.Infof("consensus: %s is too old/too new", e.String())
		return nil
	}
	if _, ok := p.Members[e.Creator]; !ok {
		p.Warnf("consensus: %s isn't member", e.Creator.String())
		return nil
	}
	id := e.Hash() // remember, because we change event here
	p.events.Add(e)
	defer p.events.DropNotFlushed()

	e.Frame, e.IsRoot = p.calcFrameIdx(e, false)
	e.MedianTime = p.events.MedianTime(id, p.PrevEpoch.Time)
	e.GasLeft = 0 // TODO
	return e
}

// checks consensus-related fields: Frame, IsRoot, MedianTimestamp, GasLeft
func (p *Poset) checkAndSaveEvent(e *inter.Event) error {
	if _, ok := p.Members[e.Creator]; !ok {
		return errors.Errorf("consensus: %s isn't member", e.Creator.String())
	}

	p.events.Add(e)
	defer p.events.DropNotFlushed()

	// check frame & isRoot
	frameIdx, isRoot := p.calcFrameIdx(e, true)
	if e.IsRoot != isRoot {
		return errors.Errorf("Claimed isRoot mismatched with calculated (%v!=%v)", e.IsRoot, isRoot)
	}
	if e.Frame != frameIdx {
		return errors.Errorf("Claimed frame mismatched with calculated (%d!=%d)", e.Frame, frameIdx)
	}
	// check median timestamp
	medianTime := p.events.MedianTime(e.Hash(), p.PrevEpoch.Time)
	if e.MedianTime != medianTime {
		return errors.Errorf("Claimed medianTime mismatched with calculated (%d!=%d)", e.MedianTime, medianTime)
	}
	// TODO check e.GasLeft

	// save in DB the {vectorindex, e}
	p.events.Flush()
	if e.IsRoot {
		p.store.AddRoot(e)
	}
	return nil
}

// calculates fiWitness election for the root, calls p.onFrameDecided if election was decided
func (p *Poset) handleElection(root *inter.Event) {
	if root != nil { // if root is nil, then just bootstrap election
		if !root.IsRoot {
			return
		}
		p.Debugf("consensus: %s is root", root.String())

		decided := p.processRoot(root.Frame, root.Creator, root.Hash())
		if decided == nil {
			return
		}

		// if we’re here, then this root has seen that lowest not decided frame is decided now
		p.onFrameDecided(decided.Frame, decided.SfWitness)
		if p.superFrameSealed(decided.SfWitness) {
			return
		}
	}

	// then call processKnownRoots until it returns nil -
	// it’s needed because new elections may already have enough votes, because we process elections from lowest to highest
	for {
		decided := p.processKnownRoots()
		if decided == nil {
			break
		}

		p.onFrameDecided(decided.Frame, decided.SfWitness)
		if p.superFrameSealed(decided.SfWitness) {
			return
		}
	}
}
func (p *Poset) processRoot(f idx.Frame, from hash.Peer, id hash.Event) (decided *election.ElectionRes) {
	decided, err := p.election.ProcessRoot(election.RootAndSlot{
		Root: id,
		Slot: election.Slot{
			Frame: f,
			Addr:  from,
		},
	})
	if err != nil {
		p.Fatal("If we're here, probably more than 1/3n are Byzantine, and the problem cannot be resolved automatically ", err)
	}
	return decided
}

// The function is similar to processRoot, but it fully re-processes the current voting.
// This routine should be called after node startup, and after each decided frame.
func (p *Poset) processKnownRoots() *election.ElectionRes {
	// iterate all the roots from LastDecidedFrame+1 to highest, call processRoot for each
	var roots []election.RootAndSlot
	p.store.ForEachRoot(p.LastDecidedFrame+1, func(f idx.Frame, from hash.Peer, id hash.Event) bool {
		roots = append(roots, election.RootAndSlot{
			Root: id,
			Slot: election.Slot{
				Frame: f,
				Addr:  from,
			},
		})
		return true
	})
	for _, root := range roots {
		decided := p.processRoot(root.Slot.Frame, root.Slot.Addr, root.Root)
		if decided != nil {
			return decided
		}
	}
	return nil
}

// consensus is not safe for concurrent use.
func (p *Poset) consensus(e *inter.Event) {
	if e.Epoch != p.SuperFrameN {
		p.Infof("consensus: %s is too old/too new", e.String())
		return
	}
	p.Debugf("consensus: start %s", e.String())

	err := p.checkAndSaveEvent(e)
	if err != nil {
		p.Warn(err)
		return
	}

	p.handleElection(e)
}

// onFrameDecided moves LastDecidedFrameN to frame.
// It includes: moving current decided frame, txs ordering and execution, superframe sealing.
func (p *Poset) onFrameDecided(frame idx.Frame, sfWitness hash.Event) {
	p.election.Reset(p.Members, frame+1)
	p.LastDecidedFrame = frame

	p.Debugf("dfsSubgraph from %s", sfWitness.String())
	unordered, err := p.dfsSubgraph(sfWitness, func(event *inter.Event) bool {
		decidedFrame := p.store.GetEventConfirmedOn(event.Hash())
		if decidedFrame == 0 {
			p.store.SetEventConfirmedOn(event.Hash(), frame)
		}
		return decidedFrame == 0
	})
	if err != nil {
		p.Fatal(err)
	}

	// ordering
	if len(unordered) == 0 {
		p.Fatal("Frame is decided with no events. It isn't possible.")
	}
	ordered := p.fareOrdering(frame, sfWitness, unordered)

	// block generation
	block := inter.NewBlock(p.checkpoint.LastBlockN+1, ordered)
	p.Debugf("block%d ordered: %s", block.Index, block.Events.String())
	p.store.SetEventsBlockNum(block.Index, ordered...)
	p.store.SetBlock(block)
	p.checkpoint.LastBlockN = block.Index
	if p.newBlockCh != nil {
		p.newBlockCh <- p.checkpoint.LastBlockN
	}

	// balances changes

	state := p.store.StateDB(p.checkpoint.Balances)
	p.applyTransactions(state, ordered, p.NextMembers)
	p.applyRewards(state, ordered, p.NextMembers)
	p.NextMembers = p.NextMembers.Top()
	balances, err := state.Commit(true)
	if err != nil {
		p.Fatal(err)
	}
	p.checkpoint.Balances = balances

	p.saveCheckpoint()
}

func (p *Poset) superFrameSealed(fiWitness hash.Event) bool {
	if p.LastDecidedFrame < SuperFrameLen {
		return false
	}

	p.nextEpoch(fiWitness)

	return true
}

func (p *Poset) getFrameRoots(f idx.Frame) EventsByPeer {
	frameRoots := EventsByPeer{}
	p.store.ForEachRoot(f, func(f idx.Frame, from hash.Peer, id hash.Event) bool {
		if f > f {
			return false
		}
		frameRoots.AddOne(id, from)
		return true
	})
	return frameRoots
}

// calcFrameIdx checks root-conditions for new event
// and returns frame where event is root.
// It is not safe for concurrent use.
func (p *Poset) calcFrameIdx(e *inter.Event, checkOnly bool) (frame idx.Frame, isRoot bool) {
	if e.SelfParent() == nil {
		// special case for first events in an SF
		frame = idx.Frame(1)
		isRoot = true
	} else {
		// calc maxParentsFrame, i.e. max(parent's frame height)
		maxParentsFrame := idx.Frame(0)
		selfParentFrame := idx.Frame(0)

		for _, parent := range e.Parents {
			pFrame := p.GetEventHeader(parent).Frame
			if maxParentsFrame == 0 || pFrame > maxParentsFrame {
				maxParentsFrame = pFrame
			}

			if e.IsSelfParent(parent) {
				selfParentFrame = pFrame
			}
		}

		// counter of all the seen roots on maxParentsFrame
		sSeenCounter := p.Members.NewCounter()
		if !checkOnly || e.IsRoot {
			// check s.seeing of prev roots only if called by creator, or if creator has marked that event is root
			for creator, roots := range p.getFrameRoots(maxParentsFrame) {
				for root := range roots {
					if p.events.StronglySee(e.Hash(), root) {
						sSeenCounter.Count(creator)
					}
				}
			}
		}
		if sSeenCounter.HasQuorum() {
			// if I see enough roots, then I become a root too
			frame = maxParentsFrame + 1
			isRoot = true
		} else {
			// I see enough roots maxParentsFrame-1, because some of my parents does. The question is - did my self-parent start the frame already?
			frame = maxParentsFrame
			isRoot = maxParentsFrame > selfParentFrame
		}
	}

	return frame, isRoot
}
