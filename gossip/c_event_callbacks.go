package gossip

import (
	"errors"
	"math/big"
	"sync/atomic"

	"github.com/Fantom-foundation/lachesis-base/gossip/dagprocessor"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-opera/eventcheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/gossip/emitter"
	"github.com/Fantom-foundation/go-opera/inter"
)

var (
	errStopped         = errors.New("service is stopped")
	errWrongMedianTime = errors.New("wrong event median time")
	errWrongEpochHash  = errors.New("wrong event epoch hash")
)

func (s *Service) buildEvent(e *inter.MutableEventPayload, onIndexed func()) error {
	// set some unique ID
	e.SetID(s.uniqueEventIDs.sample())

	// node version
	if e.Seq() <= 1 && len(s.config.Emitter.VersionToPublish) > 0 {
		version := []byte("v-" + s.config.Emitter.VersionToPublish)
		if uint32(len(version)) <= s.store.GetRules().Dag.MaxExtraData {
			e.SetExtra(version)
		}
	}

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
func (s *Service) processSavedEvent(e *inter.EventPayload, es *blockproc.EpochState) error {
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
func (s *Service) saveAndProcessEvent(e *inter.EventPayload, es *blockproc.EpochState) error {
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

// processEvent extends the engine.Process with gossip-specific actions on each event processing
func (s *Service) processEvent(e *inter.EventPayload) error {
	// s.engineMu is locked here
	if s.stopped {
		return errStopped
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

	err := s.saveAndProcessEvent(e, &es)
	if err != nil {
		return err
	}

	newEpoch := s.store.GetEpoch()

	// track events with no descendants, i.e. heads
	for _, parent := range e.Parents() {
		s.store.DelHead(oldEpoch, parent)
	}
	s.store.AddHead(oldEpoch, e.ID())
	// set validator's last event. we don't care about forks, because this index is used only for emitter
	s.store.SetLastEvent(oldEpoch, e.Creator(), e.ID())
	// update highest lamport
	if newEpoch != oldEpoch {
		s.store.SetHighestLamport(0)
	} else if e.Lamport() > s.store.GetHighestLamport() {
		s.store.SetHighestLamport(e.Lamport())
	}

	s.emitter.OnEventConnected(e)

	if newEpoch != oldEpoch {
		// reset dag indexer
		s.store.resetEpochStore(newEpoch)
		es := s.store.getEpochStore(newEpoch)
		s.dagIndexer.Reset(s.store.GetValidators(), es.table.DagIndex, func(id hash.Event) dag.Event {
			return s.store.GetEvent(id)
		})
		// notify event checkers about new validation data
		s.gasPowerCheckReader.Ctx.Store(NewGasPowerContext(s.store, s.store.GetValidators(), newEpoch, s.store.GetRules().Economy)) // read gaspower check data from disk
		s.heavyCheckReader.Addrs.Store(NewEpochPubKeys(s.store, newEpoch))
		// notify about new epoch
		s.emitter.OnNewEpoch(s.store.GetValidators(), newEpoch)
		s.feed.newEpoch.Send(newEpoch)
	}

	if s.store.IsCommitNeeded(newEpoch != oldEpoch) {
		s.blockProcWg.Wait()
		s.store.Commit()
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
	return s.pm.processor
}
