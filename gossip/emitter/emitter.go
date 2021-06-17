package emitter

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/Fantom-foundation/lachesis-base/emitter/ancestor"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	lru "github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/emitter/originatedtxs"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/tracing"
)

const (
	SenderCountBufferSize = 20000
	PayloadIndexerSize    = 5000
)

type Emitter struct {
	txTime *lru.Cache // tx hash -> tx time

	config Config

	world World

	syncStatus syncStatus

	prevIdleTime       time.Time
	prevEmittedAtTime  time.Time
	prevEmittedAtBlock idx.Block
	originatedTxs      *originatedtxs.Buffer
	pendingGas         uint64

	// note: track validators and epoch internally to avoid referring to
	// validators of a future epoch inside OnEventConnected of last epoch event
	validators *pos.Validators
	epoch      idx.Epoch

	// challenges is deadlines when each validator should emit an event
	challenges map[idx.ValidatorID]time.Time
	// offlineValidators is a map of validators which are likely to be offline
	// This map may be different on different instances
	offlineValidators     map[idx.ValidatorID]bool
	expectedEmitIntervals map[idx.ValidatorID]time.Duration
	stakeRatio            map[idx.ValidatorID]uint64

	prevRecheckedChallenges time.Time

	quorumIndexer  *ancestor.QuorumIndexer
	payloadIndexer *ancestor.PayloadIndexer

	intervals EmitIntervals

	done chan struct{}
	wg   sync.WaitGroup

	maxParents idx.Event

	logger.Periodic
}

// NewEmitter creation.
func NewEmitter(
	config Config,
	world World,
) *Emitter {
	if world.Clock == nil {
		world.Clock = &realClock{}
	}
	// Randomize event time to decrease chance of 2 parallel instances emitting event at the same time
	// It increases the chance of detecting parallel instances
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	config.EmitIntervals = config.EmitIntervals.RandomizeEmitTime(r)

	txTime, _ := lru.New(TxTimeBufferSize)
	return &Emitter{
		config:        config,
		world:         world,
		originatedTxs: originatedtxs.New(SenderCountBufferSize),
		txTime:        txTime,
		intervals:     config.EmitIntervals,
		Periodic:      logger.Periodic{Instance: logger.MakeInstance()},
	}
}

// init emitter without starting events emission
func (em *Emitter) init() {
	em.syncStatus.startup = em.world.Now()
	em.syncStatus.lastConnected = em.world.Now()
	em.syncStatus.p2pSynced = em.world.Now()
	validators, epoch := em.world.GetEpochValidators()
	em.OnNewEpoch(validators, epoch)
}

// Start starts event emission.
func (em *Emitter) Start() {
	if em.config.Validator.ID == 0 {
		// short circuit if not a validator
		return
	}
	if em.done != nil {
		return
	}
	em.init()
	em.done = make(chan struct{})

	newTxsCh := make(chan evmcore.NewTxsNotify)
	em.world.TxPool.SubscribeNewTxsNotify(newTxsCh)

	done := em.done
	em.wg.Add(1)
	go func() {
		defer em.wg.Done()
		tick := 21 * time.Millisecond
		timer := time.NewTimer(tick)
		defer timer.Stop()
		for {
			select {
			case txNotify := <-newTxsCh:
				em.memorizeTxTimes(txNotify.Txs)
			case <-timer.C:
				if isBusy := em.tick(); isBusy {
					// Heuristic to avoid locking mutexes and hurting the concurrency
					timer.Reset(tick / 3)
					continue
				}
			case <-done:
				return
			}
			timer.Reset(tick)
		}
	}()
}

// Stop stops event emission.
func (em *Emitter) Stop() {
	if em.done == nil {
		return
	}

	close(em.done)
	em.done = nil
	em.wg.Wait()
}

func (em *Emitter) tick() (isBusy bool) {
	// track synced time
	if em.world.PeersNum() == 0 {
		// connected time ~= last time when it's true that "not connected yet"
		em.syncStatus.lastConnected = em.world.Now()
	}
	if !em.world.IsSynced() {
		// synced time ~= last time when it's true that "not synced yet"
		em.syncStatus.p2pSynced = em.world.Now()
	}
	if em.world.IsBusy() {
		return true
	}

	em.recheckChallenges()
	em.recheckIdleTime()
	if em.world.Now().Sub(em.prevEmittedAtTime) >= em.intervals.Min {
		_ = em.EmitEvent()
	}

	return false
}

func (em *Emitter) EmitEvent() *inter.EventPayload {
	if em.config.Validator.ID == 0 {
		// short circuit if not a validator
		return nil
	}
	poolTxs, err := em.world.TxPool.Pending()
	if err != nil {
		em.Log.Error("Tx pool transactions fetching error", "err", err)
		return nil
	}

	if tracing.Enabled() {
		for _, tt := range poolTxs {
			for _, t := range tt {
				span := tracing.CheckTx(t.Hash(), "Emitter.EmitEvent(candidate)")
				defer span.Finish()
			}
		}
	}

	em.world.Lock()
	defer em.world.Unlock()

	start := em.world.Now()

	e := em.createEvent(poolTxs)
	if e == nil {
		return nil
	}
	em.syncStatus.prevLocalEmittedID = e.ID()

	err = em.world.Process(e)
	if err != nil {
		return nil
	}

	em.prevEmittedAtTime = em.world.Now() // record time after connecting, to add the event processing time"
	em.prevEmittedAtBlock = em.world.GetLatestBlockIndex()
	em.Log.Info("New event emitted", "id", e.ID(), "parents", len(e.Parents()), "by", e.Creator(),
		"frame", e.Frame(), "txs", e.Txs().Len(), "age", common.PrettyDuration(0), "t", common.PrettyDuration(time.Since(start)))

	// metrics
	if tracing.Enabled() {
		for _, t := range e.Txs() {
			span := tracing.CheckTx(t.Hash(), "Emitter.EmitEvent()")
			defer span.Finish()
		}
	}

	return e
}

func (em *Emitter) loadPrevEmitTime() time.Time {
	prevEventID := em.world.GetLastEvent(em.epoch, em.config.Validator.ID)
	if prevEventID == nil {
		return em.prevEmittedAtTime
	}
	prevEvent := em.world.GetEvent(*prevEventID)
	if prevEvent == nil {
		return em.prevEmittedAtTime
	}
	return prevEvent.CreationTime().Time()
}

// createEvent is not safe for concurrent use.
func (em *Emitter) createEvent(poolTxs map[common.Address]types.Transactions) *inter.EventPayload {
	if !em.isValidator() {
		return nil
	}

	if synced := em.logSyncStatus(em.isSyncedToEmit()); !synced {
		// I'm reindexing my old events, so don't create events until connect all the existing self-events
		return nil
	}

	var (
		selfParentSeq  idx.Event
		selfParentTime inter.Timestamp
		parents        hash.Events
		maxLamport     idx.Lamport
	)

	// Find parents
	selfParent, parents, ok := em.chooseParents(em.epoch, em.config.Validator.ID)
	if !ok {
		return nil
	}

	// Set parent-dependent fields
	parentHeaders := make(inter.Events, len(parents))
	for i, p := range parents {
		parent := em.world.GetEvent(p)
		if parent == nil {
			em.Log.Crit("Emitter: head not found", "mutEvent", p.String())
		}
		parentHeaders[i] = parent
		if parentHeaders[i].Creator() == em.config.Validator.ID && i != 0 {
			// there're 2 heads from me, i.e. due to a fork, chooseParents could have found multiple self-parents
			em.Periodic.Error(5*time.Second, "I've created a fork, events emitting isn't allowed", "creator", em.config.Validator.ID)
			return nil
		}
		maxLamport = idx.MaxLamport(maxLamport, parent.Lamport())
	}

	selfParentSeq = 0
	selfParentTime = 0
	var selfParentHeader *inter.Event
	if selfParent != nil {
		selfParentHeader = parentHeaders[0]
		selfParentSeq = selfParentHeader.Seq()
		selfParentTime = selfParentHeader.CreationTime()
	}

	mutEvent := &inter.MutableEventPayload{}
	mutEvent.SetEpoch(em.epoch)
	mutEvent.SetSeq(selfParentSeq + 1)
	mutEvent.SetCreator(em.config.Validator.ID)

	mutEvent.SetParents(parents)
	mutEvent.SetLamport(maxLamport + 1)
	mutEvent.SetCreationTime(inter.MaxTimestamp(inter.Timestamp(em.world.Now().UnixNano()), selfParentTime+1))

	// set consensus fields
	var metric ancestor.Metric
	err := em.world.Build(mutEvent, func() {
		// calculate event metric when it is indexed by the vector clock
		metric = eventMetric(em.quorumIndexer.GetMetricOf(mutEvent.ID()), mutEvent.Seq())
	})
	if err != nil {
		if err == ErrNotEnoughGasPower {
			em.Periodic.Warn(time.Second, "Not enough gas power to emit event. Too small stake?",
				"stake%", 100*float64(em.validators.Get(em.config.Validator.ID))/float64(em.validators.TotalWeight()))
		} else {
			em.Log.Warn("Dropped event while emitting", "err", err)
		}
		return nil
	}
	// Add txs
	em.addTxs(mutEvent, poolTxs)

	// Check if event should be emitted
	if !em.isAllowedToEmit(mutEvent, metric, selfParentHeader) {
		return nil
	}

	// calc Merkle root
	mutEvent.SetTxHash(hash.Hash(types.DeriveSha(mutEvent.Txs(), new(trie.Trie))))

	// sign
	bSig, err := em.world.Signer.Sign(em.config.Validator.PubKey, mutEvent.HashToSign().Bytes())
	if err != nil {
		em.Periodic.Error(time.Second, "Failed to sign event", "err", err)
		return nil
	}
	var sig inter.Signature
	copy(sig[:], bSig)
	mutEvent.SetSig(sig)

	// build clean event
	event := mutEvent.Build()

	// check
	if err := em.world.Check(event, parentHeaders); err != nil {
		em.Periodic.Error(time.Second, "Emitted incorrect event", "err", err)
		return nil
	}

	// set mutEvent name for debug
	em.nameEventForDebug(event)

	return event
}

func (em *Emitter) idle() bool {
	return em.originatedTxs.Empty()
}

func (em *Emitter) isValidator() bool {
	return em.config.Validator.ID != 0 && em.validators.Get(em.config.Validator.ID) != 0
}

func (em *Emitter) nameEventForDebug(e *inter.EventPayload) {
	name := []rune(hash.GetNodeName(e.Creator()))
	if len(name) < 1 {
		return
	}

	name = name[len(name)-1:]
	hash.SetEventName(e.ID(), fmt.Sprintf("%s%03d",
		strings.ToLower(string(name)),
		e.Seq()))
}
