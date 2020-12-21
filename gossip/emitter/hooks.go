package emitter

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/emitter/ancestor"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/utils/adapters/vecmt2dagidx"
)

// OnNewEpoch should be called after each epoch change, and on startup
func (em *Emitter) OnNewEpoch(newValidators *pos.Validators, newEpoch idx.Epoch) {
	if !em.isValidator() {
		return
	}
	// update myValidatorID
	em.prevEmittedAtTime = em.loadPrevEmitTime()

	em.originatedTxs.Clear()

	em.spareValidators = make(map[idx.ValidatorID]bool)
	em.offlineValidators = make(map[idx.ValidatorID]bool)
	em.challenges = make(map[idx.ValidatorID]time.Time)

	em.recountValidators(newValidators)

	em.quorumIndexer = ancestor.NewQuorumIndexer(newValidators, vecmt2dagidx.Wrap(em.world.DagIndex),
		func(median, current, update idx.Event, validatorIdx idx.Validator) ancestor.Metric {
			return updMetric(median, current, update, validatorIdx, newValidators)
		})
	em.payloadIndexer = ancestor.NewPayloadIndexer(PayloadIndexerSize)
}

// OnEventConncected tracks new events to find out am I properly synced or not
func (em *Emitter) OnEventConnected(e inter.EventPayloadI) {
	if !em.isValidator() {
		return
	}
	em.quorumIndexer.ProcessEvent(e, e.Creator() == em.config.Validator.ID)
	em.payloadIndexer.ProcessEvent(e, ancestor.Metric(e.Txs().Len()))
	for _, tx := range e.Txs() {
		em.originatedTxs.Inc(em.world.TxSender(tx))
	}
	if e.Creator() == em.config.Validator.ID && em.syncStatus.prevLocalEmittedID != e.ID() {
		// event was emitted by me on another instance
		em.onNewExternalEvent(e)
	}
	// if there was any challenge, erase it
	delete(em.challenges, e.Creator())
	// mark validator as online
	delete(em.offlineValidators, e.Creator())
}

func (em *Emitter) OnEventConfirmed(he inter.EventI) {
	if !em.isValidator() {
		return
	}
	if he.NoTxs() {
		return
	}
	e := em.world.Store.GetEventPayload(he.ID())
	for _, tx := range e.Txs() {
		em.originatedTxs.Dec(em.world.TxSender(tx))
	}
	if em.idle() {
		em.prevIdleTime = time.Now()
	}
}
