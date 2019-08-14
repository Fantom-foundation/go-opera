package gossip

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

//go:generate gencodec -type Config -formats toml -out gen_config.go

type EmitterConfig struct {
	MinEmitInterval time.Duration // minimum event emission interval
	MaxEmitInterval time.Duration // maximum event emission interval
}

type Emitter struct {
	store    *Store
	engine   Consensus
	engineMu *sync.RWMutex

	config *Config

	me         hash.Peer
	privateKey *crypto.PrivateKey

	onEmitted func(e *inter.Event)

	done chan struct{}
	wg   sync.WaitGroup
}

func NewEmitter(config *Config, me hash.Peer, privateKey *crypto.PrivateKey, engineMu *sync.RWMutex, store *Store, engine Consensus, onEmitted func(e *inter.Event)) *Emitter {
	return &Emitter{
		config:     config,
		onEmitted:  onEmitted,
		store:      store,
		me:         me,
		privateKey: privateKey,
		engine:     engine,
		engineMu:   engineMu,
	}
}

// StartEventEmission starts event emission.
func (em *Emitter) StartEventEmission() {
	if em.done != nil {
		return
	}
	em.done = make(chan struct{})

	done := em.done
	em.wg.Add(1)
	go func() {
		defer em.wg.Done()
		ticker := time.NewTicker(em.config.Emitter.MinEmitInterval)
		for {
			select {
			case <-ticker.C:
				em.EmitEvent()
			case <-done:
				return
			}
		}
	}()
}

// StopEventEmission stops event emission.
func (em *Emitter) StopEventEmission() {
	if em.done == nil {
		return
	}

	close(em.done)
	em.done = nil
	em.wg.Wait()
}

// not safe for concurrent use
func (em *Emitter) createEvent() *inter.Event {
	var (
		epoch      = em.engine.CurrentSuperFrameN()
		seq        idx.Event
		parents    hash.Events
		maxLamport idx.Lamport
	)

	posposet := em.engine.(*posposet.Poset) // TODO move FindBestParents from posposet

	selfParent, parents := posposet.FindBestParents(em.me, em.config.Dag.MaxParents, posposet.NewSeeingStrategy())

	for _, p := range parents {
		parent := em.store.GetEventHeader(p)
		if maxLamport < parent.Lamport {
			maxLamport = parent.Lamport
		}
	}

	seq = 1
	if selfParent != nil {
		seq = em.store.GetEventHeader(*selfParent).Seq + 1
	}

	event := inter.NewEvent()
	event.Epoch = epoch
	event.Seq = seq
	event.Creator = em.me
	event.Parents = parents
	event.Lamport = maxLamport + 1
	// set consensus fields
	if em.engine != nil {
		event = em.engine.Prepare(event)
		if event == nil {
			log.Warn("dropped event while emitting")
			return nil
		}
	}
	// calc hash after event is fully built
	event.RecacheHash()
	// sign
	if err := event.SignBy(em.privateKey); err != nil {
		log.Error("Failed to sign event", "err", err)
	}
	// sanity check
	if !event.VerifySignature() {
		log.Error("Produced wrong event signature")
	}

	// set event name for debug
	em.nameEventForDebug(event)

	//countEmittedEvents.Inc(1) TODO

	return event
}

func (em *Emitter) EmitEvent() *inter.Event {
	em.engineMu.Lock()
	defer em.engineMu.Unlock()

	e := em.createEvent()
	if e != nil && em.onEmitted != nil {
		em.onEmitted(e)
		log.Info("New event emitted", "e", e.String())
	}
	return e
}

func (em *Emitter) nameEventForDebug(e *inter.Event) {
	name := []rune(hash.GetNodeName(em.me))
	if len(name) < 1 {
		return
	}

	name = name[len(name)-1:]
	hash.SetEventName(e.Hash(), fmt.Sprintf("%s%03d",
		strings.ToLower(string(name)),
		e.Seq))
}
