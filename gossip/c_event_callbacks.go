package gossip

import (
	"errors"
	"math/big"
	"sync/atomic"

	"github.com/Fantom-foundation/lachesis-base/gossip/dagprocessor"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-opera/eventcheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-opera/gossip/emitter"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
	"github.com/Fantom-foundation/go-opera/utils/concurrent"
)

var (
	errStopped          = errors.New("service is stopped")
	errWrongMedianTime  = errors.New("wrong event median time")
	errWrongEpochHash   = errors.New("wrong event epoch hash")
	errNonExistingEpoch = errors.New("epoch doesn't exist")
	errSameEpoch        = errors.New("epoch hasn't changed")
	errDirtyEvmSnap     = errors.New("EVM snapshot is dirty")
)

func (s *Service) buildEvent(e *inter.MutableEventPayload, onIndexed func()) error {
	// set some unique ID
	e.SetID(s.uniqueEventIDs.sample())

	// set PrevEpochHash
	if e.Lamport() <= 1 {
		prevEpochHash := s.store.GetEpochState().Hash()
		e.SetPrevEpochHash(&prevEpochHash)
	}

	// indexing event without saving
	defer s.dagIndexer.DropNotFlushed()
	err := s.dagIndexer.Add(e)
	if err != nil {
		return err
	}

	if onIndexed != nil {
		onIndexed()
	}

	e.SetMedianTime(s.dagIndexer.MedianTime(e.ID(), s.store.GetEpochState().EpochStart))

	// calc initial GasPower
	e.SetGasPowerUsed(epochcheck.CalcGasPowerUsed(e, s.store.GetRules()))
	var selfParent *inter.Event
	if e.SelfParent() != nil {
		selfParent = s.store.GetEvent(*e.SelfParent())
	}
	availableGasPower, err := s.checkers.Gaspowercheck.CalcGasPower(e, selfParent)
	if err != nil {
		return err
	}
	if e.GasPowerUsed() > availableGasPower.Min() {
		return emitter.ErrNotEnoughGasPower
	}
	e.SetGasPowerLeft(availableGasPower.Sub(e.GasPowerUsed()))
	return s.engine.Build(e)
}

// processSavedEvent performs processing which depends on event being saved in DB
func (s *Service) processSavedEvent(e *inter.EventPayload, es *iblockproc.EpochState) error {
	err := s.dagIndexer.Add(e)
	if err != nil {
		return err
	}

	// check median time
	if e.MedianTime() != s.dagIndexer.MedianTime(e.ID(), es.EpochStart) {
		return errWrongMedianTime
	}

	// aBFT processing
	return s.engine.Process(e)
}

// saveAndProcessEvent deletes event in a case if it fails validation during event processing
func (s *Service) saveAndProcessEvent(e *inter.EventPayload, es *iblockproc.EpochState) error {
	fixEventTxHashes(e)
	// indexing event
	s.store.SetEvent(e)
	defer s.dagIndexer.DropNotFlushed()

	err := s.processSavedEvent(e, es)
	if err != nil {
		s.store.DelEvent(e.ID())
		return err
	}

	// save event index after success
	s.dagIndexer.Flush()
	return nil
}

func processEventHeads(heads *concurrent.EventsSet, e *inter.EventPayload) *concurrent.EventsSet {
	// track events with no descendants, i.e. "heads"
	heads.Lock()
	defer heads.Unlock()
	heads.Val.Erase(e.Parents()...)
	heads.Val.Add(e.ID())
	return heads
}

func processLastEvent(lasts *concurrent.ValidatorEventsSet, e *inter.EventPayload) *concurrent.ValidatorEventsSet {
	// set validator's last event. we don't care about forks, because this index is used only for emitter
	lasts.Lock()
	defer lasts.Unlock()
	lasts.Val[e.Creator()] = e.ID()
	return lasts
}

func (s *Service) switchEpochTo(newEpoch idx.Epoch) {
	s.store.SetHighestLamport(0)
	// reset dag indexer
	s.store.resetEpochStore(newEpoch)
	es := s.store.getEpochStore(newEpoch)
	s.dagIndexer.Reset(s.store.GetValidators(), es.table.DagIndex, func(id hash.Event) dag.Event {
		return s.store.GetEvent(id)
	})
	// notify event checkers about new validation data
	s.gasPowerCheckReader.Ctx.Store(NewGasPowerContext(s.store, s.store.GetValidators(), newEpoch, s.store.GetRules().Economy)) // read gaspower check data from disk
	s.heavyCheckReader.Pubkeys.Store(readEpochPubKeys(s.store, newEpoch))
	// notify about new epoch
	for _, em := range s.emitters {
		em.OnNewEpoch(s.store.GetValidators(), newEpoch)
	}
	s.feed.newEpoch.Send(newEpoch)
}

func (s *Service) SwitchEpochTo(newEpoch idx.Epoch) error {
	bs, es := s.store.GetHistoryBlockEpochState(newEpoch)
	if bs == nil {
		return errNonExistingEpoch
	}
	s.engineMu.Lock()
	defer s.engineMu.Unlock()
	s.blockProcWg.Wait()
	if newEpoch == s.store.GetEpoch() {
		return errSameEpoch
	}
	s.store.evm.RebuildEvmSnapshot(common.Hash(bs.FinalizedStateRoot))
	err := s.engine.Reset(newEpoch, es.Validators)
	if err != nil {
		return err
	}
	s.store.SetBlockEpochState(*bs, *es)
	s.switchEpochTo(newEpoch)
	s.commit(true)
	return nil
}

func (s *Service) PauseEvmSnapshot() {
	s.engineMu.Lock()
	defer s.engineMu.Unlock()
	s.blockProcWg.Wait()
	if !s.store.evm.IsEvmSnapshotPaused() {
		s.store.evm.PauseEvmSnapshot()
	}
}

func (s *Service) EvmSnapshotGeneration() bool {
	gen, _ := s.store.evm.Snaps.Generating()
	return gen
}

// processEvent extends the engine.Process with gossip-specific actions on each event processing
func (s *Service) processEvent(e *inter.EventPayload) error {
	// s.engineMu is locked here
	if s.stopped {
		return errStopped
	}
	if err := s.verWatcher.Pause(); err != nil {
		return err
	}
	if gen, err := s.store.evm.Snaps.Generating(); gen || err != nil {
		// never allow fullsync while EVM snap is still generating, as it may lead to a race condition
		s.Log.Warn("EVM snapshot is not ready during event processing", "gen", gen, "err", err)
		return errDirtyEvmSnap
	}
	atomic.StoreUint32(&s.eventBusyFlag, 1)
	defer atomic.StoreUint32(&s.eventBusyFlag, 0)

	// repeat the checks under the mutex which may depend on volatile data
	if s.store.HasEvent(e.ID()) {
		return eventcheck.ErrAlreadyConnectedEvent
	}
	if err := s.checkers.Epochcheck.Validate(e); err != nil {
		return err
	}

	oldEpoch := s.store.GetEpoch()
	es := s.store.GetEpochState()

	// check prev epoch hash
	if e.PrevEpochHash() != nil {
		if *e.PrevEpochHash() != es.Hash() {
			s.store.DelEvent(e.ID())
			return errWrongEpochHash
		}
	}

	// Process LLR votes
	err := s.processBlockVotes(inter.AsSignedBlockVotes(e))
	if err != nil && err != eventcheck.ErrAlreadyProcessedBVs {
		return err
	}
	err = s.processEpochVote(inter.AsSignedEpochVote(e))
	if err != nil && err != eventcheck.ErrAlreadyProcessedEV {
		return err
	}

	err = s.saveAndProcessEvent(e, &es)
	if err != nil {
		return err
	}

	newEpoch := s.store.GetEpoch()

	// index DAG heads and last events
	s.store.SetHeads(oldEpoch, processEventHeads(s.store.GetHeads(oldEpoch), e))
	s.store.SetLastEvents(oldEpoch, processLastEvent(s.store.GetLastEvents(oldEpoch), e))
	// update highest Lamport
	if newEpoch != oldEpoch {
		s.store.SetHighestLamport(0)
	} else if e.Lamport() > s.store.GetHighestLamport() {
		s.store.SetHighestLamport(e.Lamport())
	}

	for _, em := range s.emitters {
		em.OnEventConnected(e)
	}

	if newEpoch != oldEpoch {
		s.switchEpochTo(newEpoch)
	}

	s.mayCommit(newEpoch != oldEpoch)

	if s.haltCheck != nil && s.haltCheck(oldEpoch, newEpoch, e.MedianTime().Time()) {
		// halt syncing
		s.stopped = true
	}
	return nil
}

type uniqueID struct {
	counter *big.Int
}

func (u *uniqueID) sample() [24]byte {
	u.counter.Add(u.counter, common.Big1)
	var id [24]byte
	copy(id[:], u.counter.Bytes())
	return id
}

func (s *Service) DagProcessor() *dagprocessor.Processor {
	return s.handler.dagProcessor
}

func (s *Service) mayCommit(epochSealing bool) {
	// s.engineMu is locked here
	if epochSealing || s.store.IsCommitNeeded() {
		s.commit(epochSealing)
	}
}

func (s *Service) commit(epochSealing bool) {
	// s.engineMu is locked here
	s.blockProcWg.Wait()
	// if gcmode is full and snapsync is finalized, clean all the old state trie
	// and commit the state trie at the current block
	if !s.store.cfg.EVM.Cache.TrieDirtyDisabled && s.handler.syncStatus.AcceptEvents() {
		s.store.cleanCommitEVM()
	}
	_ = s.store.Commit()
	if epochSealing {
		s.store.CaptureEvmKvdbSnapshot()
	}
}
