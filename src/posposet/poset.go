package posposet

import (
	"math"
	"sort"
	"sync"

	"github.com/pkg/errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/election"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
)

// TODO: move it?
// using for fare ordering
var (
	prevConsensusTimestamp inter.Timestamp
	genesisTimestamp       inter.Timestamp = 1562816974
	nodeCount                              = internal.MembersCount / 3
)

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

	p.Debugf("consensus: %s is root", event.String())
	// process election for the new root
	slot := election.Slot{
		Frame: frame.Index,
		Addr:  event.Creator,
	}

	eventHash := event.Hash()

	decided, err := p.election.ProcessRoot(election.RootAndSlot{
		Root: eventHash,
		Slot: slot,
	})
	if err != nil {
		p.Fatal("Election error", err) // if we're here, probably more than 1/3n are Byzantine, and the problem cannot be resolved automatically
	}

	if decided != nil {
		// if we’re here, then this root has seen that lowest not decided frame is decided now
		p.onFrameDecided(eventHash, decided.DecidedFrame, decided.DecidedSfWitness)

		// then call processKnownRoots until it returns nil -
		// it’s needed because new elections may already have enough votes, because we process elections from lowest to highest
		for {
			decided, err := p.election.ProcessKnownRoots(p.frameNumLast(), p.getRoots)
			if err != nil {
				p.Fatal("Election error", err) // if we're here, probably more than 1/3n are Byzantine, and the problem cannot be resolved automatically
			}
			if decided != nil {
				p.onFrameDecided(eventHash, decided.DecidedFrame, decided.DecidedSfWitness)
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
func (p *Poset) onFrameDecided(eventHash hash.Event, frameDecided idx.Frame, decidedSfWitness hash.Event) {
	p.LastDecidedFrameN = frameDecided
	p.election.Reset(p.members, frameDecided+1)

	eventsToConfirm, err := p.dfsSubgraph(eventHash, p.isNotConfirmed)
	if err != nil {
		p.Fatal(err)
	}

	// ordering
	if len(eventsToConfirm) == 0 {
		return
	}

	orderedEvents := p.fareOrdering(eventsToConfirm)

	// confirm event
	for _, event := range orderedEvents {
		p.store.SetConfirmedEvent(event.Hash(), frameDecided)
	}

	// TODO: apply tx

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
			pFrame := p.FrameOfEvent(parent).Index
			if maxParentsFrame == 0 || pFrame > maxParentsFrame {
				maxParentsFrame = pFrame
			}

			if parent == e.SelfParent {
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

// Note: should be used by dfsSubgraph() as filter
func (p *Poset) isNotConfirmed(event *inter.Event) bool {
	if res := p.store.GetConfirmedEvent(event.Hash()); res == 0 {
		return true
	}
	return false
}

func (p *Poset) fareOrdering(unordered inter.Events) Events {

	// 1. Select latest events from each node with greatest lamport timestamp
	latestEvents := map[hash.Peer]*inter.Event{}
	for _, event := range unordered {
		if _, ok := latestEvents[event.Creator]; !ok {
			latestEvents[event.Creator] = event
			continue
		}

		if event.LamportTime > latestEvents[event.Creator].LamportTime {
			latestEvents[event.Creator] = event
		}
	}

	// 2. Sort by lamport
	var selectedEvents []*inter.Event
	for _, event := range latestEvents {
		selectedEvents = append(selectedEvents, event)
	}

	sort.Slice(selectedEvents, func(i, j int) bool {
		return selectedEvents[i].LamportTime < selectedEvents[j].LamportTime
	})

	if len(selectedEvents) > nodeCount {
		selectedEvents = selectedEvents[:nodeCount-1]
	}

	// 3. Get Stake from each creator
	stakes := map[hash.Peer]inter.Stake{}
	var jointStake inter.Stake
	for _, event := range selectedEvents {
		stake := p.StakeOf(event.Creator)

		stakes[event.Creator] = stake
		jointStake += stake
	}

	halfStake := jointStake / 2

	// 4. Calculate weighted median
	selectedEventsMap := map[hash.Peer]*inter.Event{}
	for _, event := range selectedEvents {
		selectedEventsMap[event.Creator] = event
	}

	var currStake inter.Stake
	var median *inter.Event
	for node, stake := range stakes {
		if currStake < halfStake {
			currStake += stake
			continue
		}

		median = selectedEventsMap[node]
		break
	}

	highestTimestamp := selectedEvents[len(selectedEvents)-1].LamportTime
	lowestTimestamp := selectedEvents[0].LamportTime

	var orderedEvents Events

	for _, event := range unordered {
		// 5. Calculate time ratio & time offset
		if prevConsensusTimestamp == 0 {
			prevConsensusTimestamp = genesisTimestamp
		}

		frameTimePeriod := math.Max(float64(median.LamportTime-prevConsensusTimestamp), 1)
		frameLamportPeriod := math.Max(float64(highestTimestamp-lowestTimestamp-1), 1)

		timeRatio := inter.Timestamp(math.Max(float64(frameTimePeriod/frameLamportPeriod), 1))

		lowestConsensusTime := prevConsensusTimestamp + timeRatio
		timeOffset := lowestConsensusTime - lowestTimestamp*timeRatio

		// 6. Calculate consensus timestamp
		consensusTimestamp := event.LamportTime*timeRatio + timeOffset
		prevConsensusTimestamp = consensusTimestamp

		orderedEvents = append(orderedEvents, &Event{event, consensusTimestamp})
	}

	sort.Sort(orderedEvents)

	return orderedEvents
}
