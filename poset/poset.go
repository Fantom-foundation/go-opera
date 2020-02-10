package poset

import (
	"github.com/pkg/errors"

	"github.com/Fantom-foundation/go-lachesis/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"github.com/Fantom-foundation/go-lachesis/logger"
	"github.com/Fantom-foundation/go-lachesis/poset/election"
	"github.com/Fantom-foundation/go-lachesis/utils"
	"github.com/Fantom-foundation/go-lachesis/vector"
)

var (
	ErrWrongEpochHash   = errors.New("mismatched prev epoch hash")
	ErrNonZeroEpochHash = errors.New("prev epoch hash isn't zero for non-first event")
	ErrCheatersObserved = errors.New("Cheaters observed by self-parent aren't allowed as parents")
	ErrWrongFrame       = errors.New("Claimed frame mismatched with calculated")
	ErrWrongIsRoot      = errors.New("Claimed isRoot mismatched with calculated")
	ErrWrongMedianTime  = errors.New("Claimed medianTime mismatched with calculated")
)

// Poset processes events to get consensus.
type Poset struct {
	dag   lachesis.DagConfig
	store *Store
	input EventSource
	*Checkpoint
	EpochState

	election *election.Election
	vecClock *vector.Index

	callback inter.ConsensusCallbacks

	epochMu utils.SpinLock // protects p.Validators and p.EpochN

	logger.Instance
}

// New creates Poset instance.
// It does not start any process.
func New(dag lachesis.DagConfig, store *Store, input EventSource) *Poset {
	p := &Poset{
		dag:   dag,
		store: store,
		input: input,

		Instance: logger.MakeInstance(),
	}

	return p
}

// GetVectorIndex returns vector clock.
func (p *Poset) GetVectorIndex() *vector.Index {
	return p.vecClock
}

// LastBlock returns current block.
func (p *Poset) LastBlock() (idx.Block, hash.Event) {
	return p.LastBlockN, p.LastAtropos
}

// Prepare fills consensus-related fields: Frame, IsRoot, MedianTimestamp, PrevEpochHash
// returns nil if event should be dropped
func (p *Poset) Prepare(e *inter.Event) *inter.Event {
	if err := epochcheck.New(&p.dag, p).Validate(e); err != nil {
		p.Log.Error("Event prepare error", "err", err, "event", e)
		return nil
	}
	id := e.Hash() // remember, because we change event here
	p.vecClock.Add(&e.EventHeaderData)
	defer p.vecClock.DropNotFlushed()

	e.Frame, e.IsRoot = p.calcFrameIdx(e, false)
	e.MedianTime = p.vecClock.MedianTime(id, p.PrevEpoch.Time)
	if e.Seq <= 1 {
		e.PrevEpochHash = p.PrevEpoch.Hash()
	} else {
		e.PrevEpochHash = hash.Zero
	}

	return e
}

// checks consensus-related fields: Frame, IsRoot, MedianTimestamp, PrevEpochHash
func (p *Poset) checkAndSaveEvent(e *inter.Event) error {
	if e.Seq <= 1 && e.PrevEpochHash != p.PrevEpoch.Hash() {
		return ErrWrongEpochHash
	}
	if e.Seq > 1 && e.PrevEpochHash != hash.Zero {
		return ErrNonZeroEpochHash
	}

	// don't link to known cheaters
	if len(p.vecClock.NoCheaters(e.SelfParent(), e.Parents)) != len(e.Parents) {
		return ErrCheatersObserved
	}

	p.vecClock.Add(&e.EventHeaderData)
	defer p.vecClock.DropNotFlushed()

	// check frame & isRoot
	frameIdx, isRoot := p.calcFrameIdx(e, true)
	if e.IsRoot != isRoot {
		return ErrWrongIsRoot
	}
	if e.Frame != frameIdx {
		return ErrWrongFrame
	}
	// check median timestamp
	medianTime := p.vecClock.MedianTime(e.Hash(), p.PrevEpoch.Time)
	if e.MedianTime != medianTime {
		return ErrWrongMedianTime
	}

	// save in DB the {vectorindex, e, heads}
	p.vecClock.Flush()
	if e.IsRoot {
		p.store.AddRoot(e)
	}

	return nil
}

// calculates Atropos election for the root, calls p.onFrameDecided if election was decided
func (p *Poset) handleElection(root *inter.Event) {
	if root != nil { // if root is nil, then just bootstrap election
		if !root.IsRoot {
			return
		}

		decided := p.processRoot(root.Frame, root.Creator, root.Hash())
		if decided == nil {
			return
		}

		// if we’re here, then this root has observed that lowest not decided frame is decided now
		if p.onFrameDecided(decided.Frame, decided.Atropos) {
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

		if p.onFrameDecided(decided.Frame, decided.Atropos) {
			return
		}
	}
}

func (p *Poset) processRoot(f idx.Frame, from idx.StakerID, id hash.Event) (decided *election.Res) {
	decided, err := p.election.ProcessRoot(election.RootAndSlot{
		ID: id,
		Slot: election.Slot{
			Frame:     f,
			Validator: from,
		},
	})
	if err != nil {
		p.Log.Crit("If we're here, probably more than 1/3W are Byzantine, and the problem cannot be resolved automatically ",
			"err", err)
	}
	return decided
}

// The function is similar to processRoot, but it fully re-processes the current voting.
// This routine should be called after node startup, and after each decided frame.
func (p *Poset) processKnownRoots() *election.Res {
	// iterate all the roots from LastDecidedFrame+1 to highest, call processRoot for each
	var decided *election.Res
	for f := p.LastDecidedFrame + 1; ; f++ {
		frameRoots := p.store.GetFrameRoots(f)
		for _, it := range frameRoots {
			p.Log.Debug("Calculate root votes in new election", "root", it.ID.String())
			decided = p.processRoot(it.Slot.Frame, it.Slot.Validator, it.ID)
			if decided != nil {
				return decided
			}
		}
		if len(frameRoots) == 0 {
			break
		}
	}
	return nil
}

// ProcessEvent takes event into processing.
// Event order matter: parents first.
// ProcessEvent is not safe for concurrent use.
func (p *Poset) ProcessEvent(e *inter.Event) (err error) {
	err = epochcheck.New(&p.dag, p).Validate(e)
	if err != nil {
		return
	}
	p.Log.Debug("Consensus: start event processing", "event", e)

	err = p.checkAndSaveEvent(e)
	if err != nil {
		return
	}

	p.handleElection(e)
	return
}

// forklessCausedByQuorumOn returns true if event is forkless caused by 2/3W roots on specified frame
func (p *Poset) forklessCausedByQuorumOn(e *inter.Event, f idx.Frame) bool {
	observedCounter := p.Validators.NewCounter()
	// check "observing" prev roots only if called by creator, or if creator has marked that event as root
	for _, it := range p.store.GetFrameRoots(f) {
		if p.vecClock.ForklessCause(e.Hash(), it.ID) {
			observedCounter.Count(it.Slot.Validator)
		}
		if observedCounter.HasQuorum() {
			break
		}
	}
	return observedCounter.HasQuorum()
}

// calcFrameIdx checks root-conditions for new event
// and returns event's frame.
// It is not safe for concurrent use.
func (p *Poset) calcFrameIdx(e *inter.Event, checkOnly bool) (frame idx.Frame, isRoot bool) {
	if len(e.Parents) == 0 {
		// special case for very first events in the epoch
		return 1, true
	}

	// calc maxParentsFrame, i.e. max(parent's frame height)
	maxParentsFrame := idx.Frame(0)
	selfParentFrame := idx.Frame(0)

	for _, parent := range e.Parents {
		pFrame := p.GetEventHeader(p.EpochN, parent).Frame
		if maxParentsFrame == 0 || pFrame > maxParentsFrame {
			maxParentsFrame = pFrame
		}

		if e.IsSelfParent(parent) {
			selfParentFrame = pFrame
		}
	}

	if checkOnly {
		// check frame & isRoot
		frame = e.Frame
		if !e.IsRoot {
			// don't check forklessCausedByQuorumOn if not claimed as root
			// if not root, then not allowed to move frame
			return selfParentFrame, false
		}
		// every root must be greater than prev. self-root. Instead, election will be faulty
		// roots aren't allowed to "jump" to higher frame than selfParentFrame+1, even if they are forkless caused
		// by 2/3W+1 there. It's because of liveness with forks, when up to 1/3W of roots on any frame may become "invisible"
		// for forklessCause relation (so if we skip frames, there's may be deadlock when frames cannot advance because there's
		// less than 2/3W visible roots)
		isRoot = frame == selfParentFrame+1 && (e.Frame <= 1 || p.forklessCausedByQuorumOn(e, e.Frame-1))
		return selfParentFrame + 1, isRoot
	}

	// calculate frame & isRoot
	if e.SelfParent() == nil {
		return 1, true
	}
	if p.forklessCausedByQuorumOn(e, selfParentFrame) {
		return selfParentFrame + 1, true
	}
	// Note: if we assign maxParentsFrame, it'll break the liveness for a case with forks, because there may be less
	// than 2/3W possible roots at maxParentsFrame, even if 1 validator is cheater and 1/3W were offline for some time
	// and didn't create roots at maxParentsFrame - they won't be able to create roots at maxParentsFrame because
	// every frame must be greater than previous
	return selfParentFrame, false

}
