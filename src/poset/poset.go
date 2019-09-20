package poset

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/Fantom-foundation/go-lachesis/src/event_check/epoch_check"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/poset/election"
	"github.com/Fantom-foundation/go-lachesis/src/vector"
)

// Poset processes events to get consensus.
type Poset struct {
	dag   lachesis.DagConfig
	store *Store
	input EventSource
	*checkpoint
	epochState

	election *election.Election
	vecClock *vector.Index

	applyBlock inter.ApplyBlockFn

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

func (p *Poset) GetVectorIndex() *vector.Index {
	return p.vecClock
}

func (p *Poset) LastBlock() (idx.Block, hash.Event) {
	return p.LastBlockN, p.LastAtropos
}

// fills consensus-related fields: Frame, IsRoot, MedianTimestamp, PrevEpochHash, GasPowerLeft
// returns nil if event should be dropped
func (p *Poset) Prepare(e *inter.Event) *inter.Event {
	if err := epoch_check.New(&p.dag, p).Validate(e); err != nil {
		p.Log.Error("Event prepare error", "err", err)
		return nil
	}
	id := e.Hash() // remember, because we change event here
	p.vecClock.Add(&e.EventHeaderData)
	defer p.vecClock.DropNotFlushed()

	e.Frame, e.IsRoot = p.calcFrameIdx(e, false)
	e.MedianTime = p.vecClock.MedianTime(id, p.PrevEpoch.Time)
	e.PrevEpochHash = p.PrevEpoch.Hash()

	gasPower := p.CalcGasPower(&e.EventHeaderData)
	if gasPower < e.GasPowerUsed {
		p.Log.Warn("Not enough gas power", "used", e.GasPowerUsed, "have", gasPower)
		return nil
	}
	e.GasPowerLeft = gasPower - e.GasPowerUsed
	return e
}

// checks consensus-related fields: Frame, IsRoot, MedianTimestamp, PrevEpochHash, GasPowerLeft
func (p *Poset) checkAndSaveEvent(e *inter.Event) error {
	if _, ok := p.Members[e.Creator]; !ok {
		return epoch_check.ErrAuth
	}
	if e.PrevEpochHash != p.PrevEpoch.Hash() {
		return errors.New("Mismatched prev epoch hash")
	}

	p.vecClock.Add(&e.EventHeaderData)
	defer p.vecClock.DropNotFlushed()

	// check frame & isRoot
	frameIdx, isRoot := p.calcFrameIdx(e, true)
	if e.IsRoot != isRoot {
		return errors.Errorf("Claimed isRoot mismatched with calculated (%v!=%v)", e.IsRoot, isRoot)
	}
	if e.Frame != frameIdx {
		return errors.Errorf("Claimed frame mismatched with calculated (%d!=%d)", e.Frame, frameIdx)
	}
	// check median timestamp
	medianTime := p.vecClock.MedianTime(e.Hash(), p.PrevEpoch.Time)
	if e.MedianTime != medianTime {
		return errors.Errorf("Claimed medianTime mismatched with calculated (%d!=%d)", e.MedianTime, medianTime)
	}
	// check GasPowerLeft
	gasPower := p.CalcGasPower(&e.EventHeaderData)
	if e.GasPowerLeft+e.GasPowerUsed != gasPower { // GasPowerUsed is checked in basic_check
		return errors.Errorf("Claimed GasPower mismatched with calculated (%d!=%d)", e.GasPowerLeft+e.GasPowerUsed, gasPower)
	}

	// save in DB the {vectorindex, e, heads}
	p.vecClock.Flush()
	if e.IsRoot {
		p.store.AddRoot(e)
	}

	return nil
}

// calculates atropos election for the root, calls p.onFrameDecided if election was decided
func (p *Poset) handleElection(root *inter.Event) {
	if root != nil { // if root is nil, then just bootstrap election
		if !root.IsRoot {
			return
		}
		p.Log.Debug("consensus: event is root", "event", root.String())

		decided := p.processRoot(root.Frame, root.Creator, root.Hash())
		if decided == nil {
			return
		}

		// if we’re here, then this root has caused that lowest not decided frame is decided now
		lastHeaders := p.onFrameDecided(decided.Frame, decided.Atropos)
		if p.tryToSealEpoch(decided.Atropos, lastHeaders) {
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

		lastHeaders := p.onFrameDecided(decided.Frame, decided.Atropos)
		if p.tryToSealEpoch(decided.Atropos, lastHeaders) {
			return
		}
	}
}
func (p *Poset) processRoot(f idx.Frame, from common.Address, id hash.Event) (decided *election.ElectionRes) {
	decided, err := p.election.ProcessRoot(election.RootAndSlot{
		Root: id,
		Slot: election.Slot{
			Frame: f,
			Addr:  from,
		},
	})
	if err != nil {
		p.Log.Crit("If we're here, probably more than 1/3n are Byzantine, and the problem cannot be resolved automatically ",
			"err", err)
	}
	return decided
}

// The function is similar to processRoot, but it fully re-processes the current voting.
// This routine should be called after node startup, and after each decided frame.
func (p *Poset) processKnownRoots() *election.ElectionRes {
	// iterate all the roots from LastDecidedFrame+1 to highest, call processRoot for each
	var decided *election.ElectionRes
	p.store.ForEachRoot(p.LastDecidedFrame+1, func(f idx.Frame, from common.Address, root hash.Event) bool {
		decided = p.processRoot(f, from, root)
		return decided == nil
	})
	return decided
}

// ProcessEvent takes event into processing.
// Event order matter: parents first.
// ProcessEvent is not safe for concurrent use.
func (p *Poset) ProcessEvent(e *inter.Event) error {
	if err := epoch_check.New(&p.dag, p).Validate(e); err != nil {
		return err
	}
	p.Log.Debug("start event processing", "event", e.String())

	err := p.checkAndSaveEvent(e)
	if err != nil {
		return err
	}

	p.handleElection(e)
	return nil
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
			pFrame := p.GetEventHeader(p.EpochN, parent).Frame
			if maxParentsFrame == 0 || pFrame > maxParentsFrame {
				maxParentsFrame = pFrame
			}

			if e.IsSelfParent(parent) {
				selfParentFrame = pFrame
			}
		}

		// counter of all the caused roots on maxParentsFrame
		forklessCausedCounter := p.Members.NewCounter()
		if !checkOnly || e.IsRoot {
			// check s.seeing of prev roots only if called by creator, or if creator has marked that event is root
			p.store.ForEachRoot(maxParentsFrame, func(f idx.Frame, from common.Address, root hash.Event) bool {
				if p.vecClock.ForklessCause(e.Hash(), root) {
					forklessCausedCounter.Count(from)
				}
				return !forklessCausedCounter.HasQuorum()
			})
		}
		if forklessCausedCounter.HasQuorum() {
			// if I cause enough roots, then I become a root too
			frame = maxParentsFrame + 1
			isRoot = true
		} else {
			// I cause enough roots maxParentsFrame-1, because some of my parents does. The question is - did my self-parent start the frame already?
			frame = maxParentsFrame
			isRoot = maxParentsFrame > selfParentFrame
		}
	}

	return frame, isRoot
}
