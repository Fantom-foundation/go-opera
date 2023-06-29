package emitter

import (
	"fmt"
	"time"

	"github.com/Fantom-foundation/lachesis-base/emitter/ancestor"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/utils/adapters/vecmt2dagidx"
	"github.com/Fantom-foundation/go-opera/version"
)

var (
	fcVersion = version.ToU64(1, 1, 3)
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

	em.recountValidators(newValidators)

	if em.switchToFCIndexer {
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

func (em *Emitter) handleVersionUpdate(e inter.EventPayloadI) {
	if e.Seq() <= 1 && len(e.Extra()) > 0 {
		var (
			vMajor int
			vMinor int
			vPatch int
			vMeta  string
		)
		n, err := fmt.Sscanf(string(e.Extra()), "v-%d.%d.%d-%s", &vMajor, &vMinor, &vPatch, &vMeta)
		if n == 4 && err == nil {
			em.validatorVersions[e.Creator()] = version.ToU64(uint16(vMajor), uint16(vMinor), uint16(vPatch))
		}
	}
}

func (em *Emitter) fcValidators() pos.Weight {
	counter := pos.Weight(0)
	for v, ver := range em.validatorVersions {
		if ver >= fcVersion {
			counter += em.validators.Get(v)
		}
	}
	return counter
}

// OnEventConnected tracks new events
func (em *Emitter) OnEventConnected(e inter.EventPayloadI) {
	em.handleVersionUpdate(e)
	if !em.switchToFCIndexer && em.fcValidators() >= pos.Weight(uint64(em.validators.TotalWeight())*5/6) {
		em.switchToFCIndexer = true
	}
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
