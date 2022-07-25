package emitter

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Fantom-foundation/lachesis-base/emitter/ancestor"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/utils/piecefunc"
	"github.com/ethereum/go-ethereum/core/types"
	lru "github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/emitter/originatedtxs"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/tracing"
	"github.com/Fantom-foundation/go-opera/utils/rate"
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

	cache struct {
		sortedTxs *types.TransactionsByPriceAndNonce
		poolTime  time.Time
		poolBlock idx.Block
		poolCount int
	}

	emittedEventFile *os.File
	emittedBvsFile   *os.File
	emittedEvFile    *os.File
	busyRate         *rate.Gauge

	logger.Periodic
}

// NewEmitter creation.
func NewEmitter(
	config Config,
	world World,
) *Emitter {
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
		Periodic:      logger.Periodic{Instance: logger.New()},
	}
}

// init emitter without starting events emission
func (em *Emitter) init() {
	em.syncStatus.startup = time.Now()
	em.syncStatus.lastConnected = time.Now()
	em.syncStatus.p2pSynced = time.Now()
	validators, epoch := em.world.GetEpochValidators()
	em.OnNewEpoch(validators, epoch)

	if len(em.config.PrevEmittedEventFile.Path) != 0 {
		em.emittedEventFile = openPrevActionFile(em.config.PrevEmittedEventFile.Path, em.config.PrevEmittedEventFile.SyncMode)
	}
	if len(em.config.PrevBlockVotesFile.Path) != 0 {
		em.emittedBvsFile = openPrevActionFile(em.config.PrevBlockVotesFile.Path, em.config.PrevBlockVotesFile.SyncMode)
	}
	if len(em.config.PrevEpochVoteFile.Path) != 0 {
		em.emittedEvFile = openPrevActionFile(em.config.PrevEpochVoteFile.Path, em.config.PrevEpochVoteFile.SyncMode)
	}
	em.busyRate = rate.NewGauge()
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
	if em.config.EmitIntervals.Min == 0 {
		return
	}
	em.wg.Add(1)
	go func() {
		defer em.wg.Done()
		tick := 11 * time.Millisecond
		timer := time.NewTimer(tick)
		defer timer.Stop()
		for {
			select {
			case txNotify := <-newTxsCh:
				em.memorizeTxTimes(txNotify.Txs)
			case <-timer.C:
				em.tick()
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
	em.busyRate.Stop()
}

func (em *Emitter) tick() {
	// track synced time
	if em.world.PeersNum() == 0 {
		// connected time ~= last time when it's true that "not connected yet"
		em.syncStatus.lastConnected = time.Now()
	}
	if !em.world.IsSynced() {
		// synced time ~= last time when it's true that "not synced yet"
		em.syncStatus.p2pSynced = time.Now()
	}
	if em.idle() {
		em.busyRate.Mark(0)
	} else {
		em.busyRate.Mark(1)
	}
	if em.world.IsBusy() {
		return
	}

	em.recheckChallenges()
	em.recheckIdleTime()
	if time.Since(em.prevEmittedAtTime) >= em.intervals.Min {
		_, _ = em.EmitEvent()
	}
}

func (em *Emitter) getSortedTxs() *types.TransactionsByPriceAndNonce {
	// Short circuit if pool wasn't updated since the cache was built
	poolCount := em.world.TxPool.Count()
	if em.cache.sortedTxs != nil &&
		em.cache.poolBlock == em.world.GetLatestBlockIndex() &&
		em.cache.poolCount == poolCount &&
		time.Since(em.cache.poolTime) < em.config.TxsCacheInvalidation {
		return em.cache.sortedTxs.Copy()
	}
	// Build the cache
	pendingTxs, err := em.world.TxPool.Pending(true)
	if err != nil {
		em.Log.Error("Tx pool transactions fetching error", "err", err)
		return nil
	}
	for from, txs := range pendingTxs {
		// Filter the excessive transactions from each sender
		if len(txs) > em.config.MaxTxsPerAddress {
			pendingTxs[from] = txs[:em.config.MaxTxsPerAddress]
		}
	}
	sortedTxs := types.NewTransactionsByPriceAndNonce(em.world.TxSigner, pendingTxs, em.world.GetRules().Economy.MinGasPrice)
	em.cache.sortedTxs = sortedTxs
	em.cache.poolCount = poolCount
	em.cache.poolBlock = em.world.GetLatestBlockIndex()
	em.cache.poolTime = time.Now()
	return sortedTxs.Copy()
}

func (em *Emitter) EmitEvent() (*inter.EventPayload, error) {
	if em.config.Validator.ID == 0 {
		// short circuit if not a validator
		return nil, nil
	}
	sortedTxs := em.getSortedTxs()

	if em.world.IsBusy() {
		return nil, nil
	}
	em.world.Lock()
	defer em.world.Unlock()

	e, err := em.createEvent(sortedTxs)
	if e == nil || err != nil {
		return nil, err
	}
	em.syncStatus.prevLocalEmittedID = e.ID()

	err = em.world.Process(e)
	if err != nil {
		em.Log.Error("Self-event connection failed", "err", err.Error())
		return nil, err
	}
	// write event ID to avoid doublesigning in future after a crash
	em.writeLastEmittedEventID(e.ID())
	if e.EpochVote().Epoch != 0 {
		em.writeLastEmittedEpochVote(e.EpochVote().Epoch)
	}
	if len(e.BlockVotes().Votes) != 0 {
		em.writeLastEmittedBlockVotes(e.BlockVotes().LastBlock())
	}
	// broadcast the event
	em.world.Broadcast(e)

	em.prevEmittedAtTime = time.Now() // record time after connecting, to add the event processing time"
	em.prevEmittedAtBlock = em.world.GetLatestBlockIndex()

	// metrics
	if tracing.Enabled() {
		for _, t := range e.Txs() {
			span := tracing.CheckTx(t.Hash(), "Emitter.EmitEvent()")
			defer span.Finish()
		}
	}

	return e, nil
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
func (em *Emitter) createEvent(sortedTxs *types.TransactionsByPriceAndNonce) (*inter.EventPayload, error) {
	if !em.isValidator() {
		return nil, nil
	}

	if synced := em.logSyncStatus(em.isSyncedToEmit()); !synced {
		// I'm reindexing my old events, so don't create events until connect all the existing self-events
		return nil, nil
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
		return nil, nil
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
			// there are 2 heads from me, i.e. due to a fork, chooseParents could have found multiple self-parents
			em.Periodic.Error(5*time.Second, "I've created a fork, events emitting isn't allowed", "creator", em.config.Validator.ID)
			return nil, nil
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

	version := uint8(0)
	if em.world.GetRules().Upgrades.Llr {
		version = 1
	}

	mutEvent := &inter.MutableEventPayload{}
	mutEvent.SetVersion(version)
	mutEvent.SetEpoch(em.epoch)
	mutEvent.SetSeq(selfParentSeq + 1)
	mutEvent.SetCreator(em.config.Validator.ID)

	mutEvent.SetParents(parents)
	mutEvent.SetLamport(maxLamport + 1)
	mutEvent.SetCreationTime(inter.MaxTimestamp(inter.Timestamp(time.Now().UnixNano()), selfParentTime+1))

	// add LLR votes
	em.addLlrEpochVote(mutEvent)
	em.addLlrBlockVotes(mutEvent)

	// node version
	if mutEvent.Seq() <= 1 && len(em.config.VersionToPublish) > 0 {
		version := []byte("v-" + em.config.VersionToPublish)
		if uint32(len(version)) <= em.world.GetRules().Dag.MaxExtraData {
			mutEvent.SetExtra(version)
		}
	}

	// set consensus fields
	var metric ancestor.Metric
	err := em.world.Build(mutEvent, func() {
		// calculate event metric when it is indexed by the vector clock
		metric = eventMetric(em.quorumIndexer.GetMetricOf(mutEvent.ID()), mutEvent.Seq())
		metric = overheadAdjustedEventMetricF(em.validators.Len(), uint64(em.busyRate.Rate1()*piecefunc.DecimalUnit), metric)
	})
	if err != nil {
		if err == ErrNotEnoughGasPower {
			em.Periodic.Warn(time.Second, "Not enough gas power to emit event. Too small stake?",
				"stake%", 100*float64(em.validators.Get(em.config.Validator.ID))/float64(em.validators.TotalWeight()))
		} else {
			em.Log.Warn("Dropped event while emitting", "err", err)
		}
		return nil, nil
	}

	// Pre-check if event should be emitted
	// It is checked in advance to avoid adding transactions just to immediately drop the event later
	if !em.isAllowedToEmit(mutEvent, true, metric, selfParentHeader) {
		return nil, nil
	}

	// Add txs
	em.addTxs(mutEvent, sortedTxs)

	// Check if event should be emitted
	// Check only if no txs were added, since check in a case with added txs was performed above
	if mutEvent.Txs().Len() == 0 {
		if !em.isAllowedToEmit(mutEvent, mutEvent.Txs().Len() != 0, metric, selfParentHeader) {
			return nil, nil
		}
	}

	// calc Payload hash
	mutEvent.SetPayloadHash(inter.CalcPayloadHash(mutEvent))

	// sign
	bSig, err := em.world.Signer.Sign(em.config.Validator.PubKey, mutEvent.HashToSign().Bytes())
	if err != nil {
		em.Periodic.Error(time.Second, "Failed to sign event", "err", err)
		return nil, err
	}
	var sig inter.Signature
	copy(sig[:], bSig)
	mutEvent.SetSig(sig)

	// build clean event
	event := mutEvent.Build()

	// check
	if err := em.world.Check(event, parentHeaders); err != nil {
		em.Periodic.Error(time.Second, "Emitted incorrect event", "err", err)
		return nil, err
	}

	// set mutEvent name for debug
	em.nameEventForDebug(event)

	return event, nil
}

func (em *Emitter) idle() bool {
	return em.originatedTxs.Empty()
}

func (em *Emitter) isValidator() bool {
	return em.config.Validator.ID != 0 && em.validators.Exists(em.config.Validator.ID)
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
