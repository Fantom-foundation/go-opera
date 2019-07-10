package posposet

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/election"
)

// stateGap is a frame-delay to apply new balance.
// TODO: move this magic number to mainnet config
const stateGap = 3

// Poset processes events to get consensus.
type Poset struct {
	store *Store
	input EventSource
	*checkpoint
	superFrame

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

func (p *Poset) getRoots(slot election.Slot) hash.Events {
	frame := p.frame(slot.Frame, false)
	if frame == nil {
		return nil
	}
	return frame.Roots[slot.Addr].Copy()
}

// rootStronglySeeRoot returns hash of root B, if root A strongly sees root B.
// Due to a fork, there may be many roots B with the same slot,
// but strongly seen may be only one of them (if no more than 1/3n are Byzantine), with a specific hash.
func (p *Poset) rootStronglySeeRoot(a hash.Event, bSlot election.Slot) *hash.Event {
	bFrame, ok := p.frames[bSlot.Frame]
	if !ok { // not known frame for B
		return nil
	}

	for b := range bFrame.Roots[bSlot.Addr] {
		if p.strongly.See(a, b) {
			return &b
		}
	}

	return nil
}

// consensus is not safe for concurrent use.
func (p *Poset) consensus(event *inter.Event) {
	p.Debugf("consensus: start %s", event.String())
	e := &Event{
		Event: event,
	}

	p.strongly.Cache(event)

	frame, isRoot := p.checkIfRoot(e)
	if !isRoot {
		return
	}

	//return // TODO: remove when election is properly worked

	p.Debugf("consensus: %s is root", event.String())
	// process election for the new root
	slot := election.Slot{
		Frame: frame.Index,
		Addr:  event.Creator,
	}

	decided, err := p.election.ProcessRoot(election.RootAndSlot{
		Root: event.Hash(),
		Slot: slot,
	})
	if err != nil {
		p.Fatal("Election error", err) // if we're here, probably more than 1/3n are Byzantine, and the problem cannot be resolved automatically
	}

	if decided != nil {
		// if we’re here, then this root has seen that lowest not decided frame is decided now
		p.onFrameDecided(decided.DecidedFrame, decided.DecidedSfWitness)

		// then call processKnownRoots until it returns nil -
		// it’s needed because new elections may already have enough votes, because we process elections from lowest to highest
		for {
			decided, err := p.election.ProcessKnownRoots(p.frameNumLast(), p.getRoots)
			if err != nil {
				p.Fatal("Election error", err) // if we're here, probably more than 1/3n are Byzantine, and the problem cannot be resolved automatically
			}
			if decided != nil {
				p.onFrameDecided(decided.DecidedFrame, decided.DecidedSfWitness)
			} else {
				break
			}
		}
	}

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

// TODO @dagchain seems like it's a handy abstranction to be called within consensus()
// moves state from frameDecided-1 to frameDecided. It includes: moving current decided frame, txs ordering and execution, superframe sealing
func (p *Poset) onFrameDecided(frameDecided idx.Frame, decidedSfWitness hash.Event) {
	p.LastFinishedFrame(frameDecided)
	p.election.Reset(p.members, p.lastFinishedFrameN+1)
	p.strongly.Reset(p.members)
}

// checkIfRoot checks root-conditions for new event
// and returns frame where event is root.
// It is not safe for concurrent use.
func (p *Poset) checkIfRoot(e *Event) (*Frame, bool) {
	var frameI idx.Frame
	isRoot := false

	if e.Index == 1 {
		// special case for first events in an SF
		frameI = idx.Frame(1)
		isRoot = true
	} else {
		// calc maxParentsFrame, i.e. max(parent's frame height)
		maxParentsFrame := idx.Frame(0)
		selfParentFrame := idx.Frame(0)

		for parent := range e.Parents {
			pEvent := p.GetEvent(parent)
			pFrame := p.FrameOfEvent(parent).Index
			if maxParentsFrame == 0 || pFrame > maxParentsFrame {
				maxParentsFrame = pFrame
			}

			if pEvent.Creator == e.Creator {
				selfParentFrame = pFrame
			}
		}

		// TODO store isRoot, frameHeight within inter.Event. Check only if event.isRoot was true.

		// counter of all the seen roots on maxParentsFrame
		sSeenCounter := p.members.NewCounter()
		for member, memberRoots := range p.frames[maxParentsFrame].Roots {
			for root := range memberRoots {
				if p.strongly.See(e.Hash(), root) {
					sSeenCounter.Count(member)
				}
			}
		}
		if sSeenCounter.HasQuorum() {
			// if I see enough roots, then I become a root too
			frameI = maxParentsFrame + 1
			isRoot = true
		} else {
			// I see enough roots maxParentsFrame-1, because some of my parents does. The question is - did my self-parent start the frame already?
			frameI = maxParentsFrame
			isRoot = maxParentsFrame > selfParentFrame
		}
	}
	// save in DB the {e, frame, isRoot}
	frame := p.frame(frameI, true)
	p.store.SetEventFrame(e.Hash(), frame.Index)
	if isRoot {
		frame.AddRoot(e)
	} else {
		frame.AddEvent(e)
	}
	return frame, isRoot
}
