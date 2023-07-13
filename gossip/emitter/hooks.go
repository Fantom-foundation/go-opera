package emitter

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/emitter/ancestor"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera/contracts/emitterdriver"
	"github.com/Fantom-foundation/go-opera/utils"
	"github.com/Fantom-foundation/go-opera/utils/adapters/vecmt2dagidx"
)

// OnNewEpoch should be called after each epoch change, and on startup
func (em *Emitter) OnNewEpoch(newValidators *pos.Validators, newEpoch idx.Epoch) {
	em.maxParents = em.config.MaxParents
	rules := em.world.GetRules()
	if em.maxParents == 0 {
		em.maxParents = rules.Dag.MaxParents
	}
	if em.maxParents > rules.Dag.MaxParents {
		em.maxParents = rules.Dag.MaxParents
	}
	if em.validators != nil && em.isValidator() && !em.validators.Exists(em.config.Validator.ID) && newValidators.Exists(em.config.Validator.ID) {
		em.syncStatus.becameValidator = time.Now()
	}

	em.validators, em.epoch = newValidators, newEpoch

	if !em.isValidator() {
		return
	}
	em.prevEmittedAtTime = em.loadPrevEmitTime()

	em.originatedTxs.Clear()
	em.pendingGas = 0

	em.offlineValidators = make(map[idx.ValidatorID]bool)
	em.challenges = make(map[idx.ValidatorID]time.Time)
	em.expectedEmitIntervals = make(map[idx.ValidatorID]time.Duration)
	em.stakeRatio = make(map[idx.ValidatorID]uint64)

	// get current adjustments from emitterdriver contract
	statedb := em.world.StateDB()
	var (
		extMinInterval        time.Duration
		extConfirmingInterval time.Duration
		switchToFCIndexer     bool
	)
	if statedb != nil {
		switchToFCIndexer = statedb.GetState(emitterdriver.ContractAddress, utils.U64to256(0)) != (common.Hash{0})
		extMinInterval = time.Duration(statedb.GetState(emitterdriver.ContractAddress, utils.U64to256(1)).Big().Uint64())
		extConfirmingInterval = time.Duration(statedb.GetState(emitterdriver.ContractAddress, utils.U64to256(2)).Big().Uint64())
	}
	if extMinInterval == 0 {
		extMinInterval = em.config.EmitIntervals.Min
	}
	if extConfirmingInterval == 0 {
		extConfirmingInterval = em.config.EmitIntervals.Confirming
	}

	// sanity check to ensure that durations aren't too small/large
	em.intervals.Min = maxDuration(minDuration(em.config.EmitIntervals.Min*20, extMinInterval), em.config.EmitIntervals.Min/4)
	em.globalConfirmingInterval = maxDuration(minDuration(em.config.EmitIntervals.Confirming*20, extConfirmingInterval), em.config.EmitIntervals.Confirming/4)
	em.recountConfirmingIntervals(newValidators)

	if switchToFCIndexer {
		em.quorumIndexer = nil
		em.fcIndexer = ancestor.NewFCIndexer(newValidators, em.world.DagIndex(), em.config.Validator.ID)
	} else {
		em.quorumIndexer = ancestor.NewQuorumIndexer(newValidators, vecmt2dagidx.Wrap(em.world.DagIndex()),
			func(median, current, update idx.Event, validatorIdx idx.Validator) ancestor.Metric {
				return updMetric(median, current, update, validatorIdx, newValidators)
			})
		em.fcIndexer = nil
	}
	em.quorumIndexer = ancestor.NewQuorumIndexer(newValidators, vecmt2dagidx.Wrap(em.world.DagIndex()),
		func(median, current, update idx.Event, validatorIdx idx.Validator) ancestor.Metric {
			return updMetric(median, current, update, validatorIdx, newValidators)
		})
	em.payloadIndexer = ancestor.NewPayloadIndexer(PayloadIndexerSize)
}

// OnEventConnected tracks new events
func (em *Emitter) OnEventConnected(e inter.EventPayloadI) {
	if !em.isValidator() {
		return
	}
	if em.fcIndexer != nil {
		em.fcIndexer.ProcessEvent(e)
	} else if em.quorumIndexer != nil {
		em.quorumIndexer.ProcessEvent(e, e.Creator() == em.config.Validator.ID)
	}
	em.payloadIndexer.ProcessEvent(e, ancestor.Metric(e.Txs().Len()))
	for _, tx := range e.Txs() {
		addr, _ := types.Sender(em.world.TxSigner, tx)
		em.originatedTxs.Inc(addr)
	}
	em.pendingGas += e.GasPowerUsed()
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
	if em.pendingGas > he.GasPowerUsed() {
		em.pendingGas -= he.GasPowerUsed()
	} else {
		em.pendingGas = 0
	}
	if he.AnyTxs() {
		e := em.world.GetEventPayload(he.ID())
		for _, tx := range e.Txs() {
			addr, _ := types.Sender(em.world.TxSigner, tx)
			em.originatedTxs.Dec(addr)
		}
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
