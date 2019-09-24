package gossip

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/params"
	"github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-lachesis/src/event_check"
	"github.com/Fantom-foundation/go-lachesis/src/event_check/basic_check"
	"github.com/Fantom-foundation/go-lachesis/src/evm_core"
	"github.com/Fantom-foundation/go-lachesis/src/gossip/occured_txs"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/ancestor"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
	"github.com/Fantom-foundation/go-lachesis/src/utils"
)

const (
	MimetypeEvent    = "application/event"
	TxTimeBufferSize = 20000
	TxTurnPeriod     = 4 * time.Second
)

type Emitter struct {
	store       *Store
	engine      Consensus
	engineMu    *sync.RWMutex
	prevEpoch   idx.Epoch
	txpool      txPool
	occurredTxs *occured_txs.Buffer
	txTime      *lru.Cache // tx hash -> tx time

	dag    *lachesis.DagConfig
	config *EmitterConfig

	am         *accounts.Manager
	coinbase   common.Address
	coinbaseMu sync.RWMutex

	gasRate         metrics.Meter
	prevEmittedTime time.Time

	onEmitted func(e *inter.Event)

	done chan struct{}
	wg   sync.WaitGroup
}

// NewEmitter creation.
func NewEmitter(
	config *Config,
	am *accounts.Manager,
	engine Consensus,
	engineMu *sync.RWMutex,
	store *Store,
	txpool txPool,
	occurredTxs *occured_txs.Buffer,
	onEmitted func(e *inter.Event),
) *Emitter {

	txTime, _ := lru.New(TxTimeBufferSize)
	return &Emitter{
		dag:         &config.Net.Dag,
		config:      &config.Emitter,
		am:          am,
		gasRate:     metrics.NewMeterForced(),
		engine:      engine,
		engineMu:    engineMu,
		store:       store,
		txpool:      txpool,
		txTime:      txTime,
		occurredTxs: occurredTxs,
		onEmitted:   onEmitted,
	}
}

// StartEventEmission starts event emission.
func (em *Emitter) StartEventEmission() {
	if em.done != nil {
		return
	}
	em.done = make(chan struct{})

	newTxsCh := make(chan evm_core.NewTxsNotify)
	em.txpool.SubscribeNewTxsNotify(newTxsCh)

	em.prevEmittedTime = em.loadPrevEmitTime()

	done := em.done
	em.wg.Add(1)
	go func() {
		defer em.wg.Done()
		ticker := time.NewTicker(em.config.MinEmitInterval / 5)
		for {
			select {
			case txNotify := <-newTxsCh:
				em.memorizeTxTimes(txNotify.Txs)
			case <-ticker.C:
				// must pass at least MinEmitInterval since last event
				if time.Since(em.prevEmittedTime) >= em.config.MinEmitInterval {
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

// SetCoinbase sets event creator.
func (em *Emitter) SetCoinbase(addr common.Address) {
	em.coinbaseMu.Lock()
	defer em.coinbaseMu.Unlock()
	em.coinbase = addr
}

// GetCoinbase gets event creator.
func (em *Emitter) GetCoinbase() common.Address {
	em.coinbaseMu.RLock()
	defer em.coinbaseMu.RUnlock()
	return em.coinbase
}

func (em *Emitter) loadPrevEmitTime() time.Time {
	prevEventId := em.store.GetLastEvent(em.engine.GetEpoch(), em.GetCoinbase())
	if prevEventId == nil {
		return em.prevEmittedTime
	}
	prevEvent := em.store.GetEventHeader(prevEventId.Epoch(), *prevEventId)
	if prevEvent == nil {
		return em.prevEmittedTime
	}
	return prevEvent.ClaimedTime.Time()
}

// safe for concurrent use
func (em *Emitter) memorizeTxTimes(txs types.Transactions) {
	now := time.Now()
	for _, tx := range txs {
		_, ok := em.txTime.Get(tx.Hash())
		if !ok {
			em.txTime.Add(tx.Hash(), now)
		}
	}
}

// safe for concurrent use
func (em *Emitter) getTxTurn(txHash common.Hash, now time.Time, membersArr []common.Address, membersArrStakes []pos.Stake) common.Address {
	var txTime time.Time
	txTimeI, ok := em.txTime.Get(txHash)
	if !ok {
		txTime = now
		em.txTime.Add(txHash, txTime)
	} else {
		txTime = txTimeI.(time.Time)
	}
	roundIndex := int((now.Sub(txTime) / TxTurnPeriod) % time.Duration(len(membersArr)))
	turn := utils.WeightedPermutation(roundIndex+1, membersArrStakes, txHash)[roundIndex]
	return membersArr[turn]
}

func (em *Emitter) addTxs(e *inter.Event) *inter.Event {
	poolTxs, err := em.txpool.Pending()
	if err != nil {
		log.Error("Tx pool transactions fetching error", "err", err)
		return e
	}

	maxGasUsed := em.maxGasPowerToUse(e)

	now := time.Now()
	members := em.engine.GetMembers()
	membersArr := members.SortedAddresses() // members must be sorted deterministically
	membersArrStakes := make([]pos.Stake, len(membersArr))
	for i, addr := range membersArr {
		membersArrStakes[i] = members[addr]
	}

	for sender, txs := range poolTxs {
		if txs.Len() > em.config.MaxTxsFromSender { // no more than MaxTxsFromSender txs from 1 sender
			txs = txs[:em.config.MaxTxsFromSender]
		}

		// txs is the chain of dependent txs
		for _, tx := range txs {
			// enough gas power
			if tx.Gas() >= e.GasPowerLeft || e.GasPowerUsed+tx.Gas() >= maxGasUsed {
				break // txs are dependent
			}
			// check not conflicted with already included txs (in any connected event)
			if em.occurredTxs.MayBeConflicted(sender, tx.Hash()) {
				break // txs are dependent
			}
			// my turn, i.e. try to not include the same tx simultaneously by different validators
			if em.getTxTurn(tx.Hash(), now, membersArr, membersArrStakes) != e.Creator {
				break // txs are dependent
			}

			// add
			e.GasPowerUsed += tx.Gas()
			e.GasPowerLeft -= tx.Gas()
			e.Transactions = append(e.Transactions, tx)
		}
	}
	// Spill txs if exceeded size limit
	// In all the "real" cases, the event will be limited by gas, not size.
	// Yet it's technically possible to construct an event which is limited by size and not by gas.
	for uint64(e.CalcSize()) > (basic_check.MaxEventSize-500) && len(e.Transactions) > 0 {
		tx := e.Transactions[len(e.Transactions)-1]
		e.GasPowerUsed -= tx.Gas()
		e.GasPowerLeft += tx.Gas()
		e.Transactions = e.Transactions[:len(e.Transactions)-1]
	}
	return e
}

func (em *Emitter) findBestParents(epoch idx.Epoch, coinbase common.Address) (*hash.Event, hash.Events, bool) {
	selfParent := em.store.GetLastEvent(epoch, coinbase)
	heads := em.store.GetHeads(epoch) // events with no descendants

	var strategy ancestor.SearchStrategy
	vecClock := em.engine.GetVectorIndex()
	if vecClock != nil {
		strategy = ancestor.NewСausalityStrategy(vecClock)

		// don't link to known cheaters
		heads = vecClock.NoCheaters(selfParent, heads)
		if selfParent != nil && len(vecClock.NoCheaters(selfParent, hash.Events{*selfParent})) == 0 {
			log.Error("I've created a fork, events emitting isn't allowed", "address", coinbase.String())
			return nil, nil, false
		}
	} else {
		// use dummy strategy in engine-less tests
		strategy = ancestor.NewRandomStrategy(nil)
	}

	_, parents := ancestor.FindBestParents(em.dag.MaxParents, heads, selfParent, strategy)
	return selfParent, parents, true
}

// createEvent is not safe for concurrent use.
func (em *Emitter) createEvent() *inter.Event {
	coinbase := em.GetCoinbase()

	if _, ok := em.engine.GetMembers()[coinbase]; !ok {
		return nil
	}

	var (
		epoch          = em.engine.GetEpoch()
		selfParentSeq  idx.Event
		selfParentTime inter.Timestamp
		parents        hash.Events
		maxLamport     idx.Lamport
	)

	// Find parents
	selfParent, parents, ok := em.findBestParents(epoch, coinbase)
	if !ok {
		return nil
	}

	// Set parent-dependent fields
	parentHeaders := make([]*inter.EventHeaderData, len(parents))
	for i, p := range parents {
		parent := em.store.GetEventHeader(epoch, p)
		if parent == nil {
			log.Crit("Emitter: head wasn't found", "e", p.String())
		}
		parentHeaders[i] = parent
		if parentHeaders[i].Creator == coinbase && i != 0 {
			// there're 2 heads from me
			log.Error("I've created a fork, events emitting isn't allowed", "address", coinbase.String())
			return nil
		}
		maxLamport = idx.MaxLamport(maxLamport, parent.Lamport)
	}

	selfParentSeq = 0
	selfParentTime = 0
	var selfParentHeader *inter.EventHeaderData
	if selfParent != nil {
		selfParentHeader = parentHeaders[0]
		selfParentSeq = selfParentHeader.Seq
		selfParentTime = selfParentHeader.ClaimedTime
	}

	event := inter.NewEvent()
	event.Epoch = epoch
	event.Seq = selfParentSeq + 1
	event.Creator = coinbase

	event.Parents = parents
	event.Lamport = maxLamport + 1
	event.ClaimedTime = inter.MaxTimestamp(inter.Timestamp(time.Now().UnixNano()), selfParentTime+1)
	event.GasPowerUsed = basic_check.CalcGasPowerUsed(event, em.dag)

	// set consensus fields
	event = em.engine.Prepare(event) // GasPowerLeft is calced here
	if event == nil {
		log.Warn("dropped event while emitting")
		return nil
	}

	// Add txs
	event = em.addTxs(event)

	if !em.isAllowedToEmit(event, selfParentHeader) {
		return nil
	}

	// calc Merkle root
	event.TxHash = types.DeriveSha(event.Transactions)

	// sign
	signer := func(data []byte) (sig []byte, err error) {
		acc := accounts.Account{
			Address: coinbase,
		}
		w, err := em.am.Find(acc)
		if err != nil {
			return
		}
		return w.SignData(acc, MimetypeEvent, data)
	}
	if err := event.Sign(signer); err != nil {
		log.Error("Failed to sign event", "err", err)
		return nil
	}
	// calc hash after event is fully built
	event.RecacheHash()
	event.RecacheSize()
	{
		// sanity check
		dagId := params.AllEthashProtocolChanges.ChainID
		if err := event_check.ValidateAll_test(em.dag, em.engine, types.NewEIP155Signer(dagId), event, parentHeaders); err != nil {
			log.Error("Emitted incorrect event", "err", err)
			return nil
		}
	}

	// set event name for debug
	em.nameEventForDebug(event)

	//TODO: countEmittedEvents.Inc(1)

	return event
}

func (em *Emitter) maxGasPowerToUse(e *inter.Event) uint64 {
	// No txs if power is low
	{
		threshold := em.config.NoTxsThreshold
		if e.GasPowerLeft <= threshold {
			return 0
		}
	}
	// Smooth TPS if power isn't big
	{
		threshold := em.config.SmoothTpsThreshold
		if e.GasPowerLeft <= threshold {
			// it's emitter, so no need in determinism => fine to use float
			passedTime := float64(e.ClaimedTime.Time().Sub(em.prevEmittedTime)) / (float64(time.Second))
			maxGasUsed := uint64(passedTime * em.gasRate.Rate1() * em.config.MaxGasRateGrowthFactor)
			if maxGasUsed > basic_check.MaxGasPowerUsed {
				maxGasUsed = basic_check.MaxGasPowerUsed
			}
			return maxGasUsed
		}
	}
	return basic_check.MaxGasPowerUsed
}

func (em *Emitter) isAllowedToEmit(e *inter.Event, selfParent *inter.EventHeaderData) bool {
	passedTime := e.ClaimedTime.Time().Sub(em.prevEmittedTime)
	// Slow down emitting if power is low
	{
		threshold := em.config.NoTxsThreshold
		if e.GasPowerLeft <= threshold {
			// it's emitter, so no need in determinism => fine to use float
			minT := float64(em.config.MinEmitInterval)
			maxT := float64(em.config.MaxEmitInterval)
			factor := float64(e.GasPowerLeft) / float64(threshold)
			adjustedEmitInterval := time.Duration(maxT - (maxT-minT)*factor)
			if passedTime < adjustedEmitInterval {
				return false
			}
		}
	}
	// Forbid emitting if not enough power and power is decreasing
	{
		threshold := em.config.EmergencyThreshold
		if e.GasPowerLeft <= threshold {
			if !(selfParent != nil && e.GasPowerLeft >= selfParent.GasPowerLeft) {
				log.Warn("Not enough power to emit event, waiting", "power", e.GasPowerLeft, "self_parent_power", selfParent.GasPowerLeft)
				return false
			}
		}
	}
	// Slow down emitting if no txs to confirm/post
	{
		if passedTime < em.config.MaxEmitInterval &&
			em.occurredTxs.Len() == 0 &&
			len(e.Transactions) == 0 {
			return false
		}
	}

	return true
}

func (em *Emitter) EmitEvent() *inter.Event {
	em.engineMu.Lock()
	defer em.engineMu.Unlock()

	e := em.createEvent()
	if e == nil {
		return nil
	}

	if em.onEmitted != nil {
		em.onEmitted(e)
	}
	em.gasRate.Mark(int64(e.GasPowerUsed))
	em.prevEmittedTime = time.Now() // record time after connecting, to add the event processing time
	log.Info("New event emitted", "e", e.String())

	return e
}

func (em *Emitter) nameEventForDebug(e *inter.Event) {
	name := []rune(hash.GetNodeName(em.coinbase))
	if len(name) < 1 {
		return
	}

	name = name[len(name)-1:]
	hash.SetEventName(e.Hash(), fmt.Sprintf("%s%03d",
		strings.ToLower(string(name)),
		e.Seq))
}
