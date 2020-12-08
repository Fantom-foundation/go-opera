package emitter

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/emitter/ancestor"
	"github.com/Fantom-foundation/lachesis-base/emitter/doublesign"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/trie"
	lru "github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/occuredtxs"
	"github.com/Fantom-foundation/go-opera/gossip/piecefunc"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/params"
	"github.com/Fantom-foundation/go-opera/tracing"
	"github.com/Fantom-foundation/go-opera/utils"
	"github.com/Fantom-foundation/go-opera/utils/adapters/vecmt2dagidx"
	"github.com/Fantom-foundation/go-opera/utils/errlock"
	"github.com/Fantom-foundation/go-opera/valkeystore"
	"github.com/Fantom-foundation/go-opera/vecmt"
)

const (
	TxTimeBufferSize = 20000
	TxTurnPeriod     = 4 * time.Second
	TxTurnNonces     = 8
)

// EmitterWorld is emitter's external world
type EmitterWorld struct {
	Store       Reader
	EngineMu    *sync.RWMutex
	Txpool      txPool
	Signer      valkeystore.SignerI
	OccurredTxs *occuredtxs.Buffer

	Check    func(e *inter.EventPayload, parents inter.Events) error
	Process  func(*inter.EventPayload) error
	Build    func(*inter.MutableEventPayload) error
	DagIndex *vecmt.Index

	IsSynced func() bool
	PeersNum func() int
}

type Emitter struct {
	txTime *lru.Cache // tx hash -> tx time

	net    opera.Rules
	config Config

	world EmitterWorld

	syncStatus syncStatus

	gasRate         metrics.Meter
	prevEmittedTime time.Time

	intervals EmitIntervals

	done chan struct{}
	wg   sync.WaitGroup

	logger.Periodic
}

type syncStatus struct {
	startup                   time.Time
	lastConnected             time.Time
	p2pSynced                 time.Time
	prevLocalEmittedID        hash.Event
	externalSelfEventCreated  time.Time
	externalSelfEventDetected time.Time
	becameValidator           time.Time
}

// NewEmitter creation.
func NewEmitter(
	net opera.Rules,
	config Config,
	world EmitterWorld,
) *Emitter {

	txTime, _ := lru.New(TxTimeBufferSize)
	loggerInstance := logger.MakeInstance()
	return &Emitter{
		net:       net,
		config:    config,
		world:     world,
		gasRate:   metrics.NewMeterForced(),
		txTime:    txTime,
		intervals: config.EmitIntervals,
		Periodic:  logger.Periodic{Instance: loggerInstance},
	}
}

// init emitter without starting events emission
func (em *Emitter) init() {
	em.syncStatus.startup = time.Now()
	em.syncStatus.lastConnected = time.Now()
	em.syncStatus.p2pSynced = time.Now()
	validators, epoch := em.world.Store.GetEpochValidators()
	em.OnNewEpoch(validators, epoch)
}

// StartEventEmission starts event emission.
func (em *Emitter) StartEventEmission() {
	if em.done != nil {
		return
	}
	em.init()
	em.done = make(chan struct{})

	newTxsCh := make(chan evmcore.NewTxsNotify)
	em.world.Txpool.SubscribeNewTxsNotify(newTxsCh)

	done := em.done
	em.wg.Add(1)
	go func() {
		defer em.wg.Done()
		ticker := time.NewTicker(10 * time.Millisecond)
		for {
			select {
			case txNotify := <-newTxsCh:
				em.memorizeTxTimes(txNotify.Txs)
			case <-ticker.C:
				// track synced time
				if em.world.PeersNum() == 0 {
					em.syncStatus.lastConnected = time.Now() // connected time ~= last time when it's true that "not connected yet"
				}
				if !em.world.IsSynced() {
					em.syncStatus.p2pSynced = time.Now() // synced time ~= last time when it's true that "not synced yet"
				}

				// must pass at least MinEmitInterval since last event
				if time.Since(em.prevEmittedTime) >= em.intervals.Min {
					em.EmitEvent()
				}
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

func (em *Emitter) loadPrevEmitTime() time.Time {
	if em.config.Validator.ID == 0 {
		return em.prevEmittedTime
	}

	prevEventID := em.world.Store.GetLastEvent(em.world.Store.GetEpoch(), em.config.Validator.ID)
	if prevEventID == nil {
		return em.prevEmittedTime
	}
	prevEvent := em.world.Store.GetEvent(*prevEventID)
	if prevEvent == nil {
		return em.prevEmittedTime
	}
	return prevEvent.CreationTime().Time()
}

// safe for concurrent use
func (em *Emitter) memorizeTxTimes(txs types.Transactions) {
	if em.config.Validator.ID == 0 {
		return // short circuit if not validator
	}
	now := time.Now()
	for _, tx := range txs {
		_, ok := em.txTime.Get(tx.Hash())
		if !ok {
			em.txTime.Add(tx.Hash(), now)
		}
	}
}

// safe for concurrent use
func (em *Emitter) isMyTxTurn(txHash common.Hash, sender common.Address, accountNonce uint64, now time.Time, validatorsArr []idx.ValidatorID, validatorsArrStakes []pos.Weight, me idx.ValidatorID, epoch idx.Epoch) bool {
	turnHash := hash.Of(sender.Bytes(), bigendian.Uint64ToBytes(accountNonce/TxTurnNonces), epoch.Bytes())

	var txTime time.Time
	txTimeI, ok := em.txTime.Get(txHash)
	if !ok {
		txTime = now
		em.txTime.Add(txHash, txTime)
	} else {
		txTime = txTimeI.(time.Time)
	}

	roundIndex := int((now.Sub(txTime) / TxTurnPeriod) % time.Duration(len(validatorsArr)))
	turns := utils.WeightedPermutation(roundIndex+1, validatorsArrStakes, turnHash)

	return validatorsArr[turns[roundIndex]] == me
}

func (em *Emitter) addTxs(e *inter.MutableEventPayload, poolTxs map[common.Address]types.Transactions) {
	if poolTxs == nil || len(poolTxs) == 0 {
		return
	}

	maxGasUsed := em.maxGasPowerToUse(e)

	now := time.Now()
	validators, epoch := em.world.Store.GetEpochValidators()
	validatorsArr := validators.SortedIDs() // validators must be sorted deterministically
	validatorsArrStakes := make([]pos.Weight, len(validatorsArr))
	for i, addr := range validatorsArr {
		validatorsArrStakes[i] = validators.Get(addr)
	}

	for sender, txs := range poolTxs {
		if txs.Len() > em.config.MaxTxsFromSender { // no more than MaxTxsFromSender txs from 1 sender
			txs = txs[:em.config.MaxTxsFromSender]
		}

		// txs is the chain of dependent txs
		for _, tx := range txs {
			// enough gas power
			if tx.Gas() >= e.GasPowerLeft().Min() || e.GasPowerUsed()+tx.Gas() >= maxGasUsed {
				break // txs are dependent, so break the loop
			}
			// check not conflicted with already included txs (in any connected event)
			if em.world.OccurredTxs.MayBeConflicted(sender, tx.Hash()) {
				break // txs are dependent, so break the loop
			}
			// my turn, i.e. try to not include the same tx simultaneously by different validators
			if !em.isMyTxTurn(tx.Hash(), sender, tx.Nonce(), now, validatorsArr, validatorsArrStakes, e.Creator(), epoch) {
				break // txs are dependent, so break the loop
			}

			// add
			e.SetGasPowerUsed(e.GasPowerUsed() + tx.Gas())
			e.SetGasPowerLeft(e.GasPowerLeft().Sub(tx.Gas()))
			e.SetTxs(append(e.Txs(), tx))
		}
	}
}

func (em *Emitter) findBestParents(epoch idx.Epoch, myValidatorID idx.ValidatorID) (*hash.Event, hash.Events, bool) {
	selfParent := em.world.Store.GetLastEvent(epoch, myValidatorID)
	heads := em.world.Store.GetHeads(epoch) // events with no descendants

	validators := em.world.Store.GetValidators()

	var strategy ancestor.SearchStrategy
	dagIndex := em.world.DagIndex
	if dagIndex != nil {
		strategy = ancestor.NewCasualityStrategy(vecmt2dagidx.Wrap(dagIndex), validators)
		if rand.Intn(20) == 0 { // every 20th event uses random strategy is avoid repeating patterns in DAG
			strategy = ancestor.NewRandomStrategy(rand.New(rand.NewSource(time.Now().UnixNano())))
		}

		// don't link to known cheaters
		heads = dagIndex.NoCheaters(selfParent, heads)
		if selfParent != nil && len(dagIndex.NoCheaters(selfParent, hash.Events{*selfParent})) == 0 {
			em.Periodic.Error(5*time.Second, "I've created a fork, events emitting isn't allowed", "creator", myValidatorID)
			return nil, nil, false
		}
	} else {
		// use dummy strategy in engine-less tests
		strategy = ancestor.NewRandomStrategy(nil)
	}

	maxParents := em.config.MaxParents
	if maxParents < em.net.Dag.MaxFreeParents {
		maxParents = em.net.Dag.MaxFreeParents
	}
	if maxParents > em.net.Dag.MaxParents {
		maxParents = em.net.Dag.MaxParents
	}
	_, parents := ancestor.FindBestParents(maxParents, heads, selfParent, strategy)
	return selfParent, parents, true
}

// createEvent is not safe for concurrent use.
func (em *Emitter) createEvent(poolTxs map[common.Address]types.Transactions) *inter.EventPayload {
	if em.config.Validator.ID == 0 {
		// not a validator
		return nil
	}

	if synced := em.logSyncStatus(em.isSyncedToEmit()); !synced {
		// I'm reindexing my old events, so don't create events until connect all the existing self-events
		return nil
	}

	var (
		epoch          = em.world.Store.GetEpoch()
		validators     = em.world.Store.GetValidators()
		selfParentSeq  idx.Event
		selfParentTime inter.Timestamp
		parents        hash.Events
		maxLamport     idx.Lamport
	)

	// Find parents
	selfParent, parents, ok := em.findBestParents(epoch, em.config.Validator.ID)
	if !ok {
		return nil
	}

	// Set parent-dependent fields
	parentHeaders := make(inter.Events, len(parents))
	for i, p := range parents {
		parent := em.world.Store.GetEvent(p)
		if parent == nil {
			em.Log.Crit("Emitter: head not found", "mutEvent", p.String())
		}
		parentHeaders[i] = parent
		if parentHeaders[i].Creator() == em.config.Validator.ID && i != 0 {
			// there're 2 heads from me, i.e. due to a fork, findBestParents could have found multiple self-parents
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
	mutEvent.SetEpoch(epoch)
	mutEvent.SetSeq(selfParentSeq + 1)
	mutEvent.SetCreator(em.config.Validator.ID)

	mutEvent.SetParents(parents)
	mutEvent.SetLamport(maxLamport + 1)
	mutEvent.SetCreationTime(inter.MaxTimestamp(inter.Timestamp(time.Now().UnixNano()), selfParentTime+inter.MinEventTime))

	// set consensus fields
	err := em.world.Build(mutEvent)
	if err != nil {
		if err == NotEnoughGasPower {
			em.Periodic.Warn(time.Second, "Not enough gas power to emit event. Too small stake?",
				"stake%", 100*float64(validators.Get(em.config.Validator.ID))/float64(validators.TotalWeight()))
		} else {
			em.Log.Warn("Dropped event while emitting", "err", err)
		}
		return nil
	}

	// Add txs
	em.addTxs(mutEvent, poolTxs)

	if !em.isAllowedToEmit(mutEvent, selfParentHeader) {
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

var (
	confirmingEmitIntervalPieces = []piecefunc.Dot{
		{
			X: 0,
			Y: 1.0 * piecefunc.DecimalUnit,
		},
		{
			X: 0.33 * piecefunc.DecimalUnit,
			Y: 1.05 * piecefunc.DecimalUnit,
		},
		{
			X: 0.66 * piecefunc.DecimalUnit,
			Y: 1.20 * piecefunc.DecimalUnit,
		},
		{
			X: 0.8 * piecefunc.DecimalUnit,
			Y: 1.5 * piecefunc.DecimalUnit,
		},
		{
			X: 0.9 * piecefunc.DecimalUnit,
			Y: 3 * piecefunc.DecimalUnit,
		},
		{
			X: 1.0 * piecefunc.DecimalUnit,
			Y: 3.9 * piecefunc.DecimalUnit,
		},
	}
	maxEmitIntervalPieces = []piecefunc.Dot{
		{
			X: 0,
			Y: 1.0 * piecefunc.DecimalUnit,
		},
		{
			X: 1.0 * piecefunc.DecimalUnit,
			Y: 0.89 * piecefunc.DecimalUnit,
		},
	}
)

// OnNewEpoch should be called after each epoch change, and on startup
func (em *Emitter) OnNewEpoch(newValidators *pos.Validators, newEpoch idx.Epoch) {
	// update myValidatorID
	em.prevEmittedTime = em.loadPrevEmitTime()

	// stakers with lower stake should emit less events to reduce network load
	// confirmingEmitInterval = piecefunc(totalStakeBeforeMe / totalStake) * MinEmitInterval
	myIdx := newValidators.GetIdx(em.config.Validator.ID)
	totalStake := pos.Weight(0)
	totalStakeBeforeMe := pos.Weight(0)
	for i, stake := range newValidators.SortedWeights() {
		totalStake += stake
		if idx.Validator(i) < myIdx {
			totalStakeBeforeMe += stake
		}
	}
	stakeRatio := uint64((totalStakeBeforeMe * piecefunc.DecimalUnit) / totalStake)
	confirmingEmitIntervalRatio := piecefunc.Get(stakeRatio, confirmingEmitIntervalPieces)
	em.intervals.Confirming = time.Duration(piecefunc.Mul(uint64(em.config.EmitIntervals.Confirming), confirmingEmitIntervalRatio))

	// stakers with lower stake should emit more events at idle, to catch up with other stakers if their frame is behind
	// MaxEmitInterval = piecefunc(totalStakeBeforeMe / totalStake) * MaxEmitInterval
	maxEmitIntervalRatio := piecefunc.Get(stakeRatio, maxEmitIntervalPieces)
	em.intervals.Max = time.Duration(piecefunc.Mul(uint64(em.config.EmitIntervals.Max), maxEmitIntervalRatio))
}

// OnNewEvent tracks new events to find out am I properly synced or not
func (em *Emitter) OnNewEvent(e *inter.EventPayload) {
	if em.config.Validator.ID == 0 || em.config.Validator.ID != e.Creator() {
		return
	}
	if em.syncStatus.prevLocalEmittedID == e.ID() {
		return
	}
	// event was emitted by me on another instance
	em.onNewExternalEvent(e)

}

func (em *Emitter) onNewExternalEvent(e *inter.EventPayload) {
	em.syncStatus.externalSelfEventDetected = time.Now()
	em.syncStatus.externalSelfEventCreated = e.CreationTime().Time()
	status := em.currentSyncStatus()
	if doublesign.DetectParallelInstance(status, em.config.EmitIntervals.ParallelInstanceProtection) {
		passedSinceEvent := status.Since(status.ExternalSelfEventCreated)
		reason := "Received a recent event (event id=%s) from this validator (validator ID=%d) which wasn't created on this node.\n" +
			"This external event was created %s, %s ago at the time of this error.\n" +
			"It might mean that a duplicating instance of the same validator is running simultaneously, which may eventually lead to a doublesign.\n" +
			"The node was stopped by one of the doublesign protection heuristics.\n" +
			"There's no guaranteed automatic protection against a doublesign," +
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
		em.Periodic.Info(25*time.Second, "Emitting is paused", "reason", syncErr)
	} else {
		em.Periodic.Info(25*time.Second, "Emitting is paused", "reason", syncErr, "wait", wait)
	}
	return false
}

// return true if event is in epoch tail (unlikely to confirm)
func (em *Emitter) isEpochTail(e dag.Event) bool {
	return e.Frame() >= idx.Frame(em.net.Dag.MaxEpochBlocks)-em.config.EpochTailLength
}

func (em *Emitter) maxGasPowerToUse(e *inter.MutableEventPayload) uint64 {
	// No txs in epoch tail, because tail events are unlikely to confirm
	{
		if em.isEpochTail(e) {
			return 0
		}
	}
	// No txs if power is low
	{
		threshold := em.config.NoTxsThreshold
		if e.GasPowerLeft().Min() <= threshold {
			return 0
		}
		if e.GasPowerLeft().Min() < threshold+params.MaxGasPowerUsed {
			return e.GasPowerLeft().Min() - threshold
		}
	}
	// Smooth TPS if power isn't big
	{
		threshold := em.config.SmoothTpsThreshold
		if e.GasPowerLeft().Min() <= threshold {
			// it's emitter, so no need in determinism => fine to use float
			passedTime := float64(e.CreationTime().Time().Sub(em.prevEmittedTime)) / (float64(time.Second))
			maxGasUsed := uint64(passedTime * em.gasRate.Rate1() * em.config.MaxGasRateGrowthFactor)
			if maxGasUsed > params.MaxGasPowerUsed {
				maxGasUsed = params.MaxGasPowerUsed
			}
			return maxGasUsed
		}
	}
	return params.MaxGasPowerUsed
}

func (em *Emitter) isAllowedToEmit(e inter.EventPayloadI, selfParent *inter.Event) bool {
	passedTime := e.CreationTime().Time().Sub(em.prevEmittedTime)
	// Slow down emitting if power is low
	{
		threshold := em.config.NoTxsThreshold
		if e.GasPowerLeft().Min() <= threshold {
			// it's emitter, so no need in determinism => fine to use float
			minT := float64(em.intervals.Min)
			maxT := float64(em.intervals.Max)
			factor := float64(e.GasPowerLeft().Min()) / float64(threshold)
			adjustedEmitInterval := time.Duration(maxT - (maxT-minT)*factor)
			if passedTime < adjustedEmitInterval {
				return false
			}
		}
	}
	// Forbid emitting if not enough power and power is decreasing
	{
		threshold := em.config.EmergencyThreshold
		if e.GasPowerLeft().Min() <= threshold {
			if selfParent != nil && e.GasPowerLeft().Min() < selfParent.GasPowerLeft().Min() {
				validators := em.world.Store.GetValidators()
				em.Periodic.Warn(10*time.Second, "Not enough power to emit event, waiting",
					"power", e.GasPowerLeft().String(),
					"selfParentPower", selfParent.GasPowerLeft().String(),
					"stake%", 100*float64(validators.Get(e.Creator()))/float64(validators.TotalWeight()))
				return false
			}
		}
	}
	// Slow down emitting if no txs to confirm/post, and not at epoch tail
	{
		if passedTime < em.intervals.Max &&
			em.world.OccurredTxs.Len() == 0 &&
			len(e.Txs()) == 0 &&
			!em.isEpochTail(e) {
			return false
		}
	}
	// Emit no more than 1 event per confirmingEmitInterval (when there's no txs to originate, but at least 1 tx to confirm)
	{
		if passedTime < em.intervals.Confirming &&
			em.world.OccurredTxs.Len() != 0 &&
			len(e.Txs()) == 0 {
			return false
		}
	}

	return true
}

func (em *Emitter) EmitEvent() *inter.EventPayload {
	if em.config.Validator.ID == 0 || em.world.Store.GetValidators().Get(em.config.Validator.ID) == 0 {
		return nil // short circuit if not validator
	}

	poolTxs, err := em.world.Txpool.Pending() // request txs before locking engineMu to prevent deadlock!
	if err != nil {
		em.Log.Error("Tx pool transactions fetching error", "err", err)
		return nil
	}

	for _, tt := range poolTxs {
		for _, t := range tt {
			span := tracing.CheckTx(t.Hash(), "Emitter.EmitEvent(candidate)")
			defer span.Finish()
		}
	}

	em.world.EngineMu.Lock()
	defer em.world.EngineMu.Unlock()

	start := time.Now()

	e := em.createEvent(poolTxs)
	if e == nil {
		return nil
	}
	em.syncStatus.prevLocalEmittedID = e.ID()

	if em.world.Process != nil {
		em.world.Process(e)
	}
	em.gasRate.Mark(int64(e.GasPowerUsed()))
	em.prevEmittedTime = time.Now() // record time after connecting, to add the event processing time"
	em.Log.Info("New event emitted", "id", e.ID(), "parents", len(e.Parents()), "by", e.Creator(), "frame", e.Frame(), "txs", e.Txs().Len(), "t", time.Since(start))

	// metrics
	for _, t := range e.Txs() {
		span := tracing.CheckTx(t.Hash(), "Emitter.EmitEvent()")
		defer span.Finish()
	}

	return e
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
