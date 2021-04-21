package emitter

import (
	"fmt"
	"time"

	"github.com/Fantom-foundation/lachesis-base/emitter/doublesign"
	"github.com/Fantom-foundation/lachesis-base/hash"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/utils/errlock"
)

type syncStatus struct {
	startup                   time.Time
	lastConnected             time.Time
	p2pSynced                 time.Time
	prevLocalEmittedID        hash.Event
	externalSelfEventCreated  time.Time
	externalSelfEventDetected time.Time
	becameValidator           time.Time
}

func (em *Emitter) onNewExternalEvent(e inter.EventPayloadI) {
	em.syncStatus.externalSelfEventDetected = time.Now()
	em.syncStatus.externalSelfEventCreated = e.CreationTime().Time()
	status := em.currentSyncStatus()
	if doublesign.DetectParallelInstance(status, em.config.EmitIntervals.ParallelInstanceProtection) {
		passedSinceEvent := status.Since(status.ExternalSelfEventCreated)
		reason := "Received a recent event (event id=%s) from this validator (validator ID=%d) which wasn't created on this node.\n" +
			"This external event was created %s, %s ago at the time of this error.\n" +
			"It might mean that a duplicating instance of the same validator is running simultaneously, which may eventually lead to a doublesign.\n" +
			"The node was stopped by one of the doublesign protection heuristics.\n" +
			"There's no guaranteed automatic protection against a doublesign, " +
			"please always ensure that no more than one instance of the same validator is running."
		errlock.Permanent(fmt.Errorf(reason, e.ID().String(), em.config.Validator.ID, e.CreationTime().Time().Local().String(), passedSinceEvent.String()))
		panic("unreachable")
	}
}

func (em *Emitter) currentSyncStatus() doublesign.SyncStatus {
	s := doublesign.SyncStatus{
		Now:                       time.Now(),
		PeersNum:                  em.world.PeersNum(),
		Startup:                   em.syncStatus.startup,
		LastConnected:             em.syncStatus.lastConnected,
		ExternalSelfEventCreated:  em.syncStatus.externalSelfEventCreated,
		ExternalSelfEventDetected: em.syncStatus.externalSelfEventDetected,
		BecameValidator:           em.syncStatus.becameValidator,
	}
	if em.world.IsSynced() {
		s.P2PSynced = em.syncStatus.p2pSynced
	}
	prevEmitted := em.readLastEmittedEventID()
	if prevEmitted != nil && (em.world.GetEvent(*prevEmitted) == nil && em.epoch <= prevEmitted.Epoch()) {
		s.P2PSynced = time.Time{}
	}
	return s
}

func (em *Emitter) isSyncedToEmit() (time.Duration, error) {
	if em.intervals.DoublesignProtection == 0 {
		return 0, nil // protection disabled
	}
	return doublesign.SyncedToEmit(em.currentSyncStatus(), em.intervals.DoublesignProtection)
}

func (em *Emitter) logSyncStatus(wait time.Duration, syncErr error) bool {
	if syncErr == nil {
		return true
	}

	if wait == 0 {
		em.Periodic.Info(7*time.Second, "Emitting is paused", "reason", syncErr)
	} else {
		em.Periodic.Info(7*time.Second, "Emitting is paused", "reason", syncErr, "wait", wait)
	}
	return false
}
