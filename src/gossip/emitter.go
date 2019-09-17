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
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/src/event_check"
	"github.com/Fantom-foundation/go-lachesis/src/event_check/basic_check"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/ancestor"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
)

const (
	MimetypeEvent = "application/event"
)

type Emitter struct {
	store     *Store
	engine    Consensus
	engineMu  *sync.RWMutex
	prevEpoch idx.Epoch
	txpool    txPool

	dag       *lachesis.DagConfig
	config    *EmitterConfig
	networkId uint64

	am         *accounts.Manager
	coinbase   common.Address
	coinbaseMu sync.RWMutex

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
	onEmitted func(e *inter.Event),
) *Emitter {

	return &Emitter{
		dag:       &config.Net.Dag,
		config:    &config.Emitter,
		am:        am,
		engine:    engine,
		engineMu:  engineMu,
		store:     store,
		txpool:    txpool,
		onEmitted: onEmitted,
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
		ticker := time.NewTicker(em.config.MinEmitInterval)
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

func (em *Emitter) addTxs(event *inter.Event) *inter.Event {
	poolTxs, err := em.txpool.Pending()
	if err != nil {
		log.Error("Tx pool transactions fetching error", "err", err)
		return event
	}
	for _, txs := range poolTxs {
		for _, tx := range txs {
			if tx.Gas() < event.GasPowerLeft && event.GasPowerUsed+tx.Gas() < basic_check.MaxGasPowerUsed {
				event.GasPowerUsed += tx.Gas()
				event.GasPowerLeft -= tx.Gas()
				event.Transactions = append(event.Transactions, txs...)
			}
		}
	}
	// Spill txs if exceeded size limit
	// In all the "real" cases, the event will be limited by gas, not size.
	// Yet it's technically possible to construct an event which is limited by size and not by gas.
	for uint64(event.CalcSize()) > basic_check.MaxEventSize && len(event.Transactions) > 0 {
		tx := event.Transactions[len(event.Transactions)-1]
		event.GasPowerUsed -= tx.Gas()
		event.GasPowerLeft += tx.Gas()
		event.Transactions = event.Transactions[:len(event.Transactions)-1]
	}
	return event
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

	vecClock := em.engine.GetVectorIndex()

	var strategy ancestor.SearchStrategy
	if vecClock != nil {
		strategy = ancestor.NewСausalityStrategy(vecClock)
	} else {
		strategy = ancestor.NewRandomStrategy(nil)
	}

	heads := em.store.GetHeads(epoch) // events with no descendants
	selfParent := em.store.GetLastEvent(epoch, coinbase)
	_, parents = ancestor.FindBestParents(em.dag.MaxParents, heads, selfParent, strategy)

	parentHeaders := make([]*inter.EventHeaderData, len(parents))
	for i, p := range parents {
		parent := em.store.GetEventHeader(epoch, p)
		if parent == nil {
			log.Crit("Emitter: head wasn't found", "e", p.String())
		}
		parentHeaders[i] = parent
		maxLamport = idx.MaxLamport(maxLamport, parent.Lamport)
	}

	selfParentSeq = 0
	selfParentTime = 0
	if selfParent != nil {
		selfParentSeq = parentHeaders[0].Seq
		selfParentTime = parentHeaders[0].ClaimedTime
	}

	event := inter.NewEvent()
	event.Epoch = epoch
	event.Seq = selfParentSeq + 1
	event.Creator = coinbase

	event.Parents = parents
	event.Lamport = maxLamport + 1
	event.ClaimedTime = inter.MaxTimestamp(inter.Timestamp(time.Now().UnixNano()), selfParentTime+1)
	event.GasPowerUsed = basic_check.CalcGasPowerUsed(event)

	// set consensus fields
	event = em.engine.Prepare(event) // GasPowerLeft is calced here
	if event == nil {
		log.Warn("dropped event while emitting")
		return nil
	}

	// Add txs
	event = em.addTxs(event)

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
	name := []rune(hash.GetNodeName(em.coinbase))
	if len(name) < 1 {
		return
	}

	name = name[len(name)-1:]
	hash.SetEventName(e.Hash(), fmt.Sprintf("%s%03d",
		strings.ToLower(string(name)),
		e.Seq))
}
