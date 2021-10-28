package gossip

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Fantom-foundation/lachesis-base/gossip/dagprocessor"
	"github.com/Fantom-foundation/lachesis-base/gossip/itemsfetcher"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/utils/datasemaphore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/eventcheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/bvallcheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/evallcheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/heavycheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/parentlesscheck"
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/blockrecords/brprocessor"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/blockrecords/brstream"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/blockrecords/brstream/brstreamleecher"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/blockrecords/brstream/brstreamseeder"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/blockvotes/bvprocessor"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/blockvotes/bvstream"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/blockvotes/bvstream/bvstreamleecher"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/blockvotes/bvstream/bvstreamseeder"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/dag/dagstream"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/dag/dagstream/dagstreamleecher"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/dag/dagstream/dagstreamseeder"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/epochpacks/epprocessor"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/epochpacks/epstream"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/epochpacks/epstream/epstreamleecher"
	"github.com/Fantom-foundation/go-opera/gossip/protocols/epochpacks/epstream/epstreamseeder"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"
	"github.com/Fantom-foundation/go-opera/logger"
)

const (
	softResponseLimitSize = 2 * 1024 * 1024    // Target maximum size of returned events, or other data.
	softLimitItems        = 250                // Target maximum number of events or transactions to request/response
	hardLimitItems        = softLimitItems * 4 // Maximum number of events or transactions to request/response

	// txChanSize is the size of channel listening to NewTxsNotify.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096
)

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

func checkLenLimits(size int, v interface{}) error {
	if size <= 0 {
		return errResp(ErrEmptyMessage, "%v", v)
	}
	if size > hardLimitItems {
		return errResp(ErrMsgTooLarge, "%v", v)
	}
	return nil
}

type dagNotifier interface {
	SubscribeNewEpoch(ch chan<- idx.Epoch) notify.Subscription
	SubscribeNewEmitted(ch chan<- *inter.EventPayload) notify.Subscription
}

type processCallback struct {
	Event func(*inter.EventPayload) error
	BVs   func(inter.LlrSignedBlockVotes) error
	BR    func(ibr.LlrIdxFullBlockRecord) error
	EV    func(inter.LlrSignedEpochVote) error
	ER    func(ier.LlrIdxFullEpochRecord) error
}

// handlerConfig is the collection of initialization parameters to create a full
// node network handler.
type handlerConfig struct {
	config   Config
	notifier dagNotifier
	txpool   txPool
	engineMu sync.Locker
	checkers *eventcheck.Checkers
	s        *Store
	process  processCallback
}

type ProtocolManager struct {
	NetworkID uint64
	config    Config

	synced uint32 // Flag whether we're considered synchronised (enables transaction processing, events broadcasting)

	txpool   txPool
	maxPeers int

	peers *peerSet

	txsCh  chan evmcore.NewTxsNotify
	txsSub notify.Subscription

	dagLeecher   *dagstreamleecher.Leecher
	dagSeeder    *dagstreamseeder.Seeder
	dagProcessor *dagprocessor.Processor
	dagFetcher   *itemsfetcher.Fetcher

	bvLeecher   *bvstreamleecher.Leecher
	bvSeeder    *bvstreamseeder.Seeder
	bvProcessor *bvprocessor.Processor

	brLeecher   *brstreamleecher.Leecher
	brSeeder    *brstreamseeder.Seeder
	brProcessor *brprocessor.Processor

	epLeecher   *epstreamleecher.Leecher
	epSeeder    *epstreamseeder.Seeder
	epProcessor *epprocessor.Processor

	process processCallback

	txFetcher *itemsfetcher.Fetcher

	checkers *eventcheck.Checkers

	msgSemaphore *datasemaphore.DataSemaphore

	store    *Store
	engineMu sync.Locker

	notifier             dagNotifier
	emittedEventsCh      chan *inter.EventPayload
	emittedEventsSub     notify.Subscription
	newEpochsCh          chan idx.Epoch
	newEpochsSub         notify.Subscription
	quitProgressBradcast chan struct{}

	// channels for syncer, txsyncLoop
	newPeerCh   chan *peer
	txsyncCh    chan *txsync
	quitSync    chan struct{}
	noMorePeers chan struct{}

	// wait group is used for graceful shutdowns during downloading
	// and processing
	loopsWg sync.WaitGroup
	wg      sync.WaitGroup

	logger.Instance
}

// newHandler returns a new Fantom sub protocol manager. The Fantom sub protocol manages peers capable
// with the Fantom network.
func newHandler(
	c handlerConfig,
) (
	*ProtocolManager,
	error,
) {
	// Create the protocol manager with the base fields
	pm := &ProtocolManager{
		NetworkID:            c.s.GetRules().NetworkID,
		config:               c.config,
		notifier:             c.notifier,
		txpool:               c.txpool,
		msgSemaphore:         datasemaphore.New(c.config.Protocol.MsgsSemaphoreLimit, getSemaphoreWarningFn("P2P messages")),
		store:                c.s,
		process:              c.process,
		checkers:             c.checkers,
		peers:                newPeerSet(),
		engineMu:             c.engineMu,
		newPeerCh:            make(chan *peer),
		noMorePeers:          make(chan struct{}),
		txsyncCh:             make(chan *txsync),
		quitSync:             make(chan struct{}),
		quitProgressBradcast: make(chan struct{}),

		Instance: logger.New("PM"),
	}

	pm.dagFetcher = itemsfetcher.New(pm.config.Protocol.DagFetcher, itemsfetcher.Callback{
		OnlyInterested: func(ids []interface{}) []interface{} {
			return pm.onlyInterestedEventsI(ids)
		},
		Suspend: func() bool {
			return false
		},
	})
	pm.txFetcher = itemsfetcher.New(pm.config.Protocol.TxFetcher, itemsfetcher.Callback{
		OnlyInterested: func(txids []interface{}) []interface{} {
			return txidsToInterfaces(pm.txpool.OnlyNotExisting(interfacesToTxids(txids)))
		},
		Suspend: func() bool {
			return false
		},
	})

	pm.dagProcessor = pm.makeDagProcessor(c.checkers)
	pm.dagLeecher = dagstreamleecher.New(pm.store.GetEpoch(), pm.store.GetHighestLamport() == 0, pm.config.Protocol.DagStreamLeecher, dagstreamleecher.Callbacks{
		IsProcessed: pm.store.HasEvent,
		RequestChunk: func(peer string, r dagstream.Request) error {
			p := pm.peers.Peer(peer)
			if p == nil {
				return errNotRegistered
			}
			return p.RequestEventsStream(r)
		},
		Suspend: func(_ string) bool {
			return pm.dagFetcher.Overloaded() || pm.dagProcessor.Overloaded()
		},
		PeerEpoch: func(peer string) idx.Epoch {
			p := pm.peers.Peer(peer)
			if p == nil {
				return 0
			}
			return p.progress.Epoch
		},
	})
	pm.dagSeeder = dagstreamseeder.New(pm.config.Protocol.DagStreamSeeder, dagstreamseeder.Callbacks{
		ForEachEvent: c.s.ForEachEventRLP,
	})

	pm.bvProcessor = pm.makeBvProcessor(c.checkers)
	pm.bvLeecher = bvstreamleecher.New(pm.config.Protocol.BvStreamLeecher, bvstreamleecher.Callbacks{
		LowestBlockToDecide: func() (idx.Epoch, idx.Block) {
			llrs := pm.store.GetLlrState()
			epoch := pm.store.FindBlockEpoch(llrs.LowestBlockToDecide)
			return epoch, llrs.LowestBlockToDecide
		},
		MaxEpochToDecide: func() idx.Epoch {
			return pm.store.GetLlrState().LowestEpochToFill
		},
		IsProcessed: pm.store.HasBlockVotes,
		RequestChunk: func(peer string, r bvstream.Request) error {
			p := pm.peers.Peer(peer)
			if p == nil {
				return errNotRegistered
			}
			return p.RequestBVsStream(r)
		},
		Suspend: func(_ string) bool {
			return pm.bvProcessor.Overloaded()
		},
		PeerBlock: func(peer string) idx.Block {
			p := pm.peers.Peer(peer)
			if p == nil {
				return 0
			}
			return p.progress.LastBlockIdx
		},
	})
	pm.bvSeeder = bvstreamseeder.New(pm.config.Protocol.BvStreamSeeder, bvstreamseeder.Callbacks{
		Iterate: pm.store.IterateOverlappingBlockVotesRLP,
	})

	pm.brProcessor = pm.makeBrProcessor()
	pm.brLeecher = brstreamleecher.New(pm.config.Protocol.BrStreamLeecher, brstreamleecher.Callbacks{
		LowestBlockToFill: func() idx.Block {
			return pm.store.GetLlrState().LowestBlockToFill
		},
		MaxBlockToFill: func() idx.Block {
			// rough estimation for the max fill-able block
			llrs := pm.store.GetLlrState()
			start := llrs.LowestBlockToFill
			end := llrs.LowestBlockToDecide
			if end > start+100 && pm.store.HasBlock(start+100) {
				return start + 100
			}
			return end
		},
		IsProcessed: pm.store.HasBlock,
		RequestChunk: func(peer string, r brstream.Request) error {
			p := pm.peers.Peer(peer)
			if p == nil {
				return errNotRegistered
			}
			return p.RequestBRsStream(r)
		},
		Suspend: func(_ string) bool {
			return pm.brProcessor.Overloaded()
		},
		PeerBlock: func(peer string) idx.Block {
			p := pm.peers.Peer(peer)
			if p == nil {
				return 0
			}
			return p.progress.LastBlockIdx
		},
	})
	pm.brSeeder = brstreamseeder.New(pm.config.Protocol.BrStreamSeeder, brstreamseeder.Callbacks{
		Iterate: pm.store.IterateFullBlockRecordsRLP,
	})

	pm.epProcessor = pm.makeEpProcessor(pm.checkers)
	pm.epLeecher = epstreamleecher.New(pm.config.Protocol.EpStreamLeecher, epstreamleecher.Callbacks{
		LowestEpochToFetch: func() idx.Epoch {
			return pm.store.GetLlrState().LowestEpochToFill
		},
		MaxEpochToFetch: func() idx.Epoch {
			return pm.store.GetLlrState().LowestEpochToDecide + 10000
		},
		IsProcessed: pm.store.HasHistoryBlockEpochState,
		RequestChunk: func(peer string, r epstream.Request) error {
			p := pm.peers.Peer(peer)
			if p == nil {
				return errNotRegistered
			}
			return p.RequestEPsStream(r)
		},
		Suspend: func(_ string) bool {
			return pm.epProcessor.Overloaded()
		},
		PeerEpoch: func(peer string) idx.Epoch {
			p := pm.peers.Peer(peer)
			if p == nil {
				return 0
			}
			return p.progress.Epoch
		},
	})
	pm.epSeeder = epstreamseeder.New(pm.config.Protocol.EpStreamSeeder, epstreamseeder.Callbacks{
		Iterate: pm.store.IterateEpochPacksRLP,
	})

	return pm, nil
}

func (pm *ProtocolManager) peerMisbehaviour(peer string, err error) bool {
	if eventcheck.IsBan(err) {
		log.Warn("Dropping peer due to a misbehaviour", "peer", peer, "err", err)
		pm.removePeer(peer)
		return true
	}
	return false
}

func (pm *ProtocolManager) makeDagProcessor(checkers *eventcheck.Checkers) *dagprocessor.Processor {
	// checkers
	lightCheck := func(e dag.Event) error {
		if pm.store.GetEpoch() != e.ID().Epoch() {
			return epochcheck.ErrNotRelevant
		}
		if pm.dagProcessor.IsBuffered(e.ID()) {
			return eventcheck.ErrDuplicateEvent
		}
		if pm.store.HasEvent(e.ID()) {
			return eventcheck.ErrAlreadyConnectedEvent
		}
		if err := checkers.Basiccheck.Validate(e.(inter.EventPayloadI)); err != nil {
			return err
		}
		if err := checkers.Epochcheck.Validate(e.(inter.EventPayloadI)); err != nil {
			return err
		}
		return nil
	}
	bufferedCheck := func(_e dag.Event, _parents dag.Events) error {
		e := _e.(inter.EventI)
		parents := make(inter.EventIs, len(_parents))
		for i := range _parents {
			parents[i] = _parents[i].(inter.EventI)
		}
		var selfParent inter.EventI
		if e.SelfParent() != nil {
			selfParent = parents[0].(inter.EventI)
		}
		if err := checkers.Parentscheck.Validate(e, parents); err != nil {
			return err
		}
		if err := checkers.Gaspowercheck.Validate(e, selfParent); err != nil {
			return err
		}
		return nil
	}
	parentlessChecker := parentlesscheck.Checker{
		HeavyCheck: &heavycheck.EventsOnly{Checker: checkers.Heavycheck},
		LightCheck: lightCheck,
	}
	newProcessor := dagprocessor.New(datasemaphore.New(pm.config.Protocol.EventsSemaphoreLimit, getSemaphoreWarningFn("DAG events")), pm.config.Protocol.DagProcessor, dagprocessor.Callback{
		// DAG callbacks
		Event: dagprocessor.EventCallback{
			Process: func(_e dag.Event) error {
				e := _e.(*inter.EventPayload)
				preStart := time.Now()
				pm.engineMu.Lock()
				defer pm.engineMu.Unlock()

				err := pm.process.Event(e)
				if err != nil {
					return err
				}

				// event is connected, announce it if synced up
				if atomic.LoadUint32(&pm.synced) != 0 {
					passedSinceEvent := preStart.Sub(e.CreationTime().Time())
					pm.BroadcastEvent(e, passedSinceEvent)
				}

				return nil
			},
			Released: func(e dag.Event, peer string, err error) {
				if eventcheck.IsBan(err) {
					log.Warn("Incoming event rejected", "event", e.ID().String(), "creator", e.Creator(), "err", err)
					pm.removePeer(peer)
				}
			},

			Exists: func(id hash.Event) bool {
				return pm.store.HasEvent(id)
			},

			Get: func(id hash.Event) dag.Event {
				e := pm.store.GetEventPayload(id)
				if e == nil {
					return nil
				}
				return e
			},

			CheckParents:    bufferedCheck,
			CheckParentless: parentlessChecker.Enqueue,
		},
		HighestLamport: pm.store.GetHighestLamport,
	})

	return newProcessor
}

func (pm *ProtocolManager) makeBvProcessor(checkers *eventcheck.Checkers) *bvprocessor.Processor {
	// checkers
	lightCheck := func(bvs inter.LlrSignedBlockVotes) error {
		if pm.store.HasBlockVotes(bvs.Epoch, bvs.LastBlock(), bvs.EventLocator.ID()) {
			return eventcheck.ErrAlreadyProcessedBVs
		}
		return checkers.Basiccheck.ValidateBVs(bvs)
	}
	allChecker := bvallcheck.Checker{
		HeavyCheck: &heavycheck.BVsOnly{Checker: checkers.Heavycheck},
		LightCheck: lightCheck,
	}
	return bvprocessor.New(datasemaphore.New(pm.config.Protocol.BVsSemaphoreLimit, getSemaphoreWarningFn("BVs")), pm.config.Protocol.BvProcessor, bvprocessor.Callback{
		// DAG callbacks
		Item: bvprocessor.ItemCallback{
			Process: func(bvs inter.LlrSignedBlockVotes) error {
				pm.engineMu.Lock()
				defer pm.engineMu.Unlock()
				return pm.process.BVs(bvs)
			},
			Released: func(bvs inter.LlrSignedBlockVotes, peer string, err error) {
				if eventcheck.IsBan(err) {
					log.Warn("Incoming BVs rejected", "BVs", bvs.EventLocator.ID(), "creator", bvs.EventLocator.Creator, "err", err)
					pm.removePeer(peer)
				}
			},
			Check: allChecker.Enqueue,
		},
	})
}

func (pm *ProtocolManager) makeBrProcessor() *brprocessor.Processor {
	// checkers
	return brprocessor.New(datasemaphore.New(pm.config.Protocol.BVsSemaphoreLimit, getSemaphoreWarningFn("BR")), pm.config.Protocol.BrProcessor, brprocessor.Callback{
		// DAG callbacks
		Item: brprocessor.ItemCallback{
			Process: pm.process.BR,
			Released: func(br ibr.LlrIdxFullBlockRecord, peer string, err error) {
				if eventcheck.IsBan(err) {
					log.Warn("Incoming BR rejected", "block", br.Idx, "err", err)
					pm.removePeer(peer)
				}
			},
		},
	})
}

func (pm *ProtocolManager) makeEpProcessor(checkers *eventcheck.Checkers) *epprocessor.Processor {
	// checkers
	lightCheck := func(ev inter.LlrSignedEpochVote) error {
		if pm.store.HasEpochVote(ev.Epoch, ev.EventLocator.ID()) {
			return eventcheck.ErrAlreadyProcessedEV
		}
		return checkers.Basiccheck.ValidateEV(ev)
	}
	allChecker := evallcheck.Checker{
		HeavyCheck: &heavycheck.EVOnly{Checker: checkers.Heavycheck},
		LightCheck: lightCheck,
	}
	// checkers
	return epprocessor.New(datasemaphore.New(pm.config.Protocol.BVsSemaphoreLimit, getSemaphoreWarningFn("BR")), pm.config.Protocol.EpProcessor, epprocessor.Callback{
		// DAG callbacks
		Item: epprocessor.ItemCallback{
			ProcessEV: func(ev inter.LlrSignedEpochVote) error {
				pm.engineMu.Lock()
				defer pm.engineMu.Unlock()
				return pm.process.EV(ev)
			},
			ProcessER: pm.process.ER,
			ReleasedEV: func(ev inter.LlrSignedEpochVote, peer string, err error) {
				if eventcheck.IsBan(err) {
					log.Warn("Incoming EV rejected", "event", ev.EventLocator.ID(), "creator", ev.EventLocator.Creator, "err", err)
					pm.removePeer(peer)
				}
			},
			ReleasedER: func(er ier.LlrIdxFullEpochRecord, peer string, err error) {
				if eventcheck.IsBan(err) {
					log.Warn("Incoming ER rejected", "epoch", er.Idx, "err", err)
					pm.removePeer(peer)
				}
			},
			CheckEV: allChecker.Enqueue,
		},
	})
}

func (pm *ProtocolManager) isEventInterested(id hash.Event, epoch idx.Epoch) bool {
	if id.Epoch() != epoch {
		return false
	}
	if pm.dagProcessor.IsBuffered(id) || pm.store.HasEvent(id) {
		return false
	}
	return true
}

func (pm *ProtocolManager) onlyInterestedEventsI(ids []interface{}) []interface{} {
	if len(ids) == 0 {
		return ids
	}
	epoch := pm.store.GetEpoch()
	interested := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		if pm.isEventInterested(id.(hash.Event), epoch) {
			interested = append(interested, id)
		}
	}
	return interested
}

func (pm *ProtocolManager) removePeer(id string) {
	peer := pm.peers.Peer(id)
	if peer != nil {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
	}
}

func (pm *ProtocolManager) unregisterPeer(id string) {
	// Short circuit if the peer was already removed
	peer := pm.peers.Peer(id)
	if peer == nil {
		return
	}
	log.Debug("Removing peer", "peer", id)

	// Unregister the peer from the leecher's and seeder's and peer sets
	_ = pm.epLeecher.UnregisterPeer(id)
	_ = pm.epSeeder.UnregisterPeer(id)
	_ = pm.dagLeecher.UnregisterPeer(id)
	_ = pm.dagSeeder.UnregisterPeer(id)
	_ = pm.brLeecher.UnregisterPeer(id)
	_ = pm.brSeeder.UnregisterPeer(id)
	_ = pm.bvLeecher.UnregisterPeer(id)
	_ = pm.bvSeeder.UnregisterPeer(id)
	if err := pm.peers.Unregister(id); err != nil {
		log.Error("Peer removal failed", "peer", id, "err", err)
	}
}

func (pm *ProtocolManager) Start(maxPeers int) {
	pm.maxPeers = maxPeers

	// broadcast transactions
	pm.txsCh = make(chan evmcore.NewTxsNotify, txChanSize)
	pm.txsSub = pm.txpool.SubscribeNewTxsNotify(pm.txsCh)

	pm.loopsWg.Add(1)
	go pm.txBroadcastLoop()

	if pm.notifier != nil {
		// broadcast mined events
		pm.emittedEventsCh = make(chan *inter.EventPayload, 4)
		pm.emittedEventsSub = pm.notifier.SubscribeNewEmitted(pm.emittedEventsCh)
		// epoch changes
		pm.newEpochsCh = make(chan idx.Epoch, 4)
		pm.newEpochsSub = pm.notifier.SubscribeNewEpoch(pm.newEpochsCh)

		pm.loopsWg.Add(3)
		go pm.emittedBroadcastLoop()
		go pm.progressBroadcastLoop()
		go pm.onNewEpochLoop()
	}

	// start sync handlers
	go pm.syncer()
	go pm.txsyncLoop()
	pm.dagFetcher.Start()
	pm.txFetcher.Start()
	pm.checkers.Heavycheck.Start()

	pm.epProcessor.Start()
	pm.epSeeder.Start()
	pm.epLeecher.Start()

	pm.dagProcessor.Start()
	pm.dagSeeder.Start()
	pm.dagLeecher.Start()

	pm.bvProcessor.Start()
	pm.bvSeeder.Start()
	pm.bvLeecher.Start()

	pm.brProcessor.Start()
	pm.brSeeder.Start()
	pm.brLeecher.Start()
}

func (pm *ProtocolManager) Stop() {
	log.Info("Stopping Fantom protocol")

	pm.brLeecher.Stop()
	pm.brSeeder.Stop()
	pm.brProcessor.Stop()

	pm.bvLeecher.Stop()
	pm.bvSeeder.Stop()
	pm.bvProcessor.Stop()

	pm.dagLeecher.Stop()
	pm.dagSeeder.Stop()
	pm.dagProcessor.Stop()

	pm.epLeecher.Start()
	pm.epSeeder.Start()
	pm.epProcessor.Start()

	pm.checkers.Heavycheck.Stop()
	pm.txFetcher.Stop()
	pm.dagFetcher.Stop()

	close(pm.quitProgressBradcast)
	pm.txsSub.Unsubscribe() // quits txBroadcastLoop
	if pm.notifier != nil {
		pm.emittedEventsSub.Unsubscribe() // quits eventBroadcastLoop
		pm.newEpochsSub.Unsubscribe()     // quits onNewEpochLoop
	}

	// Wait for the subscription loops to come down.
	pm.loopsWg.Wait()

	pm.msgSemaphore.Terminate()
	// Quit the sync loop.
	// After this send has completed, no new peers will be accepted.
	pm.noMorePeers <- struct{}{}
	close(pm.quitSync)

	// Disconnect existing sessions.
	// This also closes the gate for any new registrations on the peer set.
	// sessions which are already established but not added to pm.peers yet
	// will exit when they try to register.
	pm.peers.Close()

	// Wait for all peer handler goroutines to come down.
	pm.wg.Wait()

	log.Info("Fantom protocol stopped")
}

func (pm *ProtocolManager) myProgress() PeerProgress {
	bs := pm.store.GetBlockState()
	epoch := pm.store.GetEpoch()
	return PeerProgress{
		Epoch:            epoch,
		LastBlockIdx:     bs.LastBlock.Idx,
		LastBlockAtropos: bs.LastBlock.Atropos,
	}
}

func (pm *ProtocolManager) highestPeerProgress() PeerProgress {
	peers := pm.peers.List()
	max := pm.myProgress()
	for _, peer := range peers {
		if max.LastBlockIdx < peer.progress.LastBlockIdx {
			max = peer.progress
		}
	}
	return max
}

// handle is the callback invoked to manage the life cycle of a peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p *peer) error {
	// Ignore maxPeers if this is a trusted peer
	if pm.peers.Len() >= pm.maxPeers && !p.Peer.Info().Network.Trusted {
		return p2p.DiscTooManyPeers
	}
	p.Log().Debug("Peer connected", "name", p.Name())

	// Execute the handshake
	var (
		genesis    = *pm.store.GetGenesisHash()
		myProgress = pm.myProgress()
	)
	if err := p.Handshake(pm.NetworkID, myProgress, common.Hash(genesis)); err != nil {
		p.Log().Debug("Handshake failed", "err", err)
		return err
	}
	//if rw, ok := p.rw.(*meteredMsgReadWriter); ok {
	//	rw.Init(p.version)
	//}
	// Register the peer locally
	if err := pm.peers.Register(p); err != nil {
		p.Log().Warn("Peer registration failed", "err", err)
		return err
	}
	if err := pm.dagLeecher.RegisterPeer(p.id); err != nil {
		p.Log().Warn("Leecher peer registration failed", "err", err)
		return err
	}
	if p.RunningCap(ProtocolName, []uint{FTM63}) {
		if err := pm.epLeecher.RegisterPeer(p.id); err != nil {
			p.Log().Warn("Leecher peer registration failed", "err", err)
			return err
		}
		if err := pm.bvLeecher.RegisterPeer(p.id); err != nil {
			p.Log().Warn("Leecher peer registration failed", "err", err)
			return err
		}
		if err := pm.brLeecher.RegisterPeer(p.id); err != nil {
			p.Log().Warn("Leecher peer registration failed", "err", err)
			return err
		}
	}
	defer pm.unregisterPeer(p.id)

	// Propagate existing transactions. new transactions appearing
	// after this will be sent via broadcasts.
	pm.syncTransactions(p, pm.txpool.SampleHashes(pm.config.Protocol.MaxInitialTxHashesSend))

	// Handle incoming messages until the connection is torn down
	for {
		if err := pm.handleMsg(p); err != nil {
			p.Log().Debug("Message handling failed", "err", err)
			return err
		}
	}
}

func interfacesToEventIDs(ids []interface{}) hash.Events {
	res := make(hash.Events, len(ids))
	for i, id := range ids {
		res[i] = id.(hash.Event)
	}
	return res
}

func eventIDsToInterfaces(ids hash.Events) []interface{} {
	res := make([]interface{}, len(ids))
	for i, id := range ids {
		res[i] = id
	}
	return res
}

func interfacesToTxids(ids []interface{}) []common.Hash {
	res := make([]common.Hash, len(ids))
	for i, id := range ids {
		res[i] = id.(common.Hash)
	}
	return res
}

func txidsToInterfaces(ids []common.Hash) []interface{} {
	res := make([]interface{}, len(ids))
	for i, id := range ids {
		res[i] = id
	}
	return res
}

func (pm *ProtocolManager) handleTxHashes(p *peer, announces []common.Hash) {
	// Mark the hashes as present at the remote node
	for _, id := range announces {
		p.MarkTransaction(id)
	}
	// Schedule all the unknown hashes for retrieval
	requestTransactions := func(ids []interface{}) error {
		return p.RequestTransactions(interfacesToTxids(ids))
	}
	_ = pm.txFetcher.NotifyAnnounces(p.id, txidsToInterfaces(announces), time.Now(), requestTransactions)
}

func (pm *ProtocolManager) handleTxs(p *peer, txs types.Transactions) {
	// Mark the hashes as present at the remote node
	for _, tx := range txs {
		p.MarkTransaction(tx.Hash())
	}
	pm.txpool.AddRemotes(txs)
}

func (pm *ProtocolManager) handleEventHashes(p *peer, announces hash.Events) {
	// Mark the hashes as present at the remote node
	for _, id := range announces {
		p.MarkEvent(id)
	}
	// filter too high IDs
	notTooHigh := make(hash.Events, 0, len(announces))
	sessionCfg := pm.config.Protocol.DagStreamLeecher.Session
	for _, id := range announces {
		maxLamport := pm.store.GetHighestLamport() + idx.Lamport(sessionCfg.DefaultChunkItemsNum+1)*idx.Lamport(sessionCfg.ParallelChunksDownload)
		if id.Lamport() <= maxLamport {
			notTooHigh = append(notTooHigh, id)
		}
	}
	if len(announces) != len(notTooHigh) {
		pm.dagLeecher.ForceSyncing()
	}
	if len(notTooHigh) == 0 {
		return
	}
	// Schedule all the unknown hashes for retrieval
	requestEvents := func(ids []interface{}) error {
		return p.RequestEvents(interfacesToEventIDs(ids))
	}
	_ = pm.dagFetcher.NotifyAnnounces(p.id, eventIDsToInterfaces(notTooHigh), time.Now(), requestEvents)
}

func (pm *ProtocolManager) handleEvents(p *peer, events dag.Events, ordered bool) {
	// Mark the hashes as present at the remote node
	for _, e := range events {
		p.MarkEvent(e.ID())
	}
	// filter too high events
	notTooHigh := make(dag.Events, 0, len(events))
	sessionCfg := pm.config.Protocol.DagStreamLeecher.Session
	for _, e := range events {
		maxLamport := pm.store.GetHighestLamport() + idx.Lamport(sessionCfg.DefaultChunkItemsNum+1)*idx.Lamport(sessionCfg.ParallelChunksDownload)
		if e.Lamport() <= maxLamport {
			notTooHigh = append(notTooHigh, e)
		}
	}
	if len(events) != len(notTooHigh) {
		pm.dagLeecher.ForceSyncing()
	}
	if len(notTooHigh) == 0 {
		return
	}
	// Schedule all the events for connection
	peer := *p
	now := time.Now()
	requestEvents := func(ids []interface{}) error {
		return peer.RequestEvents(interfacesToEventIDs(ids))
	}
	notifyAnnounces := func(ids hash.Events) {
		_ = pm.dagFetcher.NotifyAnnounces(peer.id, eventIDsToInterfaces(ids), now, requestEvents)
	}
	_ = pm.dagProcessor.Enqueue(peer.id, notTooHigh, ordered, notifyAnnounces, nil)
}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (pm *ProtocolManager) handleMsg(p *peer) error {
	// Read the next message from the remote peer, and ensure it's fully consumed
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Size > protocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, protocolMaxMsgSize)
	}
	defer msg.Discard()
	// Acquire semaphore for serialized messages
	eventsSizeEst := dag.Metric{
		Num:  1,
		Size: uint64(msg.Size),
	}
	if !pm.msgSemaphore.Acquire(eventsSizeEst, pm.config.Protocol.MsgsSemaphoreTimeout) {
		pm.Log.Warn("Failed to acquire semaphore for p2p message", "size", msg.Size, "peer", p.id)
		return nil
	}
	defer pm.msgSemaphore.Release(eventsSizeEst)

	myEpoch := pm.store.GetEpoch()

	// Handle the message depending on its contents
	switch {
	case msg.Code == HandshakeMsg:
		// Status messages should never arrive after the handshake
		return errResp(ErrExtraStatusMsg, "uncontrolled status message")

	case msg.Code == ProgressMsg:
		var progress PeerProgress
		if err := msg.Decode(&progress); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		p.SetProgress(progress)
		if progress.Epoch == myEpoch {
			atomic.StoreUint32(&pm.synced, 1) // Mark initial sync done on any peer which has the same epoch
		}

	case msg.Code == EvmTxsMsg:
		// Transactions arrived, make sure we have a valid and fresh graph to handle them
		if atomic.LoadUint32(&pm.synced) == 0 {
			break
		}
		// Transactions can be processed, parse all of them and deliver to the pool
		var txs types.Transactions
		if err := msg.Decode(&txs); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		if err := checkLenLimits(len(txs), txs); err != nil {
			return err
		}
		txids := make([]interface{}, txs.Len())
		for i, tx := range txs {
			txids[i] = tx.Hash()
		}
		_ = pm.txFetcher.NotifyReceived(txids)
		pm.handleTxs(p, txs)

	case msg.Code == NewEvmTxHashesMsg:
		// Transactions arrived, make sure we have a valid and fresh graph to handle them
		if atomic.LoadUint32(&pm.synced) == 0 {
			break
		}
		// Transactions can be processed, parse all of them and deliver to the pool
		var txHashes []common.Hash
		if err := msg.Decode(&txHashes); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		if err := checkLenLimits(len(txHashes), txHashes); err != nil {
			return err
		}
		pm.handleTxHashes(p, txHashes)

	case msg.Code == GetEvmTxsMsg:
		// Transactions can be processed, parse all of them and deliver to the pool
		var requests []common.Hash
		if err := msg.Decode(&requests); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		if err := checkLenLimits(len(requests), requests); err != nil {
			return err
		}

		txs := make(types.Transactions, 0, len(requests))
		for _, txid := range requests {
			tx := pm.txpool.Get(txid)
			if tx == nil {
				continue
			}
			txs = append(txs, tx)
		}
		SplitTransactions(txs, func(batch types.Transactions) {
			p.EnqueueSendTransactions(batch, p.queue)
		})

	case msg.Code == EventsMsg:
		var events inter.EventPayloads
		if err := msg.Decode(&events); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if err := checkLenLimits(len(events), events); err != nil {
			return err
		}
		_ = pm.dagFetcher.NotifyReceived(eventIDsToInterfaces(events.IDs()))
		pm.handleEvents(p, events.Bases(), events.Len() > 1)

	case msg.Code == NewEventIDsMsg:
		// Fresh events arrived, make sure we have a valid and fresh graph to handle them
		if atomic.LoadUint32(&pm.synced) == 0 {
			break
		}
		var announces hash.Events
		if err := msg.Decode(&announces); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if err := checkLenLimits(len(announces), announces); err != nil {
			return err
		}
		pm.handleEventHashes(p, announces)

	case msg.Code == GetEventsMsg:
		var requests hash.Events
		if err := msg.Decode(&requests); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if err := checkLenLimits(len(requests), requests); err != nil {
			return err
		}

		rawEvents := make([]rlp.RawValue, 0, len(requests))
		ids := make(hash.Events, 0, len(requests))
		size := 0
		for _, id := range requests {
			if raw := pm.store.GetEventPayloadRLP(id); raw != nil {
				rawEvents = append(rawEvents, raw)
				ids = append(ids, id)
				size += len(raw)
			} else {
				pm.Log.Debug("requested event not found", "hash", id)
			}
			if size >= softResponseLimitSize {
				break
			}
		}
		if len(rawEvents) != 0 {
			p.EnqueueSendEventsRLP(rawEvents, ids, p.queue)
		}

	case msg.Code == RequestEventsStream:
		var request dagstream.Request
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if request.Limit.Num > hardLimitItems-1 {
			return errResp(ErrMsgTooLarge, "%v", msg)
		}
		if request.Limit.Size > protocolMaxMsgSize*2/3 {
			return errResp(ErrMsgTooLarge, "%v", msg)
		}

		pid := p.id
		_, peerErr := pm.dagSeeder.NotifyRequestReceived(dagstreamseeder.Peer{
			ID:        pid,
			SendChunk: p.SendEventsStream,
			Misbehaviour: func(err error) {
				pm.peerMisbehaviour(pid, err)
			},
		}, request)
		if peerErr != nil {
			return peerErr
		}

	case msg.Code == EventsStreamResponse:
		var chunk dagChunk
		if err := msg.Decode(&chunk); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if err := checkLenLimits(len(chunk.Events)+len(chunk.IDs)+1, chunk); err != nil {
			return err
		}

		if (len(chunk.Events) != 0) && (len(chunk.IDs) != 0) {
			return errors.New("expected either events or event hashes")
		}
		var last hash.Event
		if len(chunk.IDs) != 0 {
			pm.handleEventHashes(p, chunk.IDs)
			last = chunk.IDs[len(chunk.IDs)-1]
		}
		if len(chunk.Events) != 0 {
			pm.handleEvents(p, chunk.Events.Bases(), true)
			last = chunk.Events[len(chunk.Events)-1].ID()
		}

		_ = pm.dagLeecher.NotifyChunkReceived(chunk.SessionID, last, chunk.Done)

	case msg.Code == RequestBVsStream:
		var request bvstream.Request
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if request.Limit.Num > hardLimitItems-1 {
			return errResp(ErrMsgTooLarge, "%v", msg)
		}
		if request.Limit.Size > protocolMaxMsgSize*2/3 {
			return errResp(ErrMsgTooLarge, "%v", msg)
		}

		pid := p.id
		_, peerErr := pm.bvSeeder.NotifyRequestReceived(bvstreamseeder.Peer{
			ID:        pid,
			SendChunk: p.SendBVsStream,
			Misbehaviour: func(err error) {
				pm.peerMisbehaviour(pid, err)
			},
		}, request)
		if peerErr != nil {
			return peerErr
		}

	case msg.Code == BVsStreamResponse:
		var chunk bvsChunk
		if err := msg.Decode(&chunk); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if err := checkLenLimits(len(chunk.BVs)+1, chunk); err != nil {
			return err
		}

		var last bvstreamleecher.BVsID
		if len(chunk.BVs) != 0 {
			_ = pm.bvProcessor.Enqueue(p.id, chunk.BVs, nil)
			last = bvstreamleecher.BVsID{
				Epoch:     chunk.BVs[len(chunk.BVs)-1].Epoch,
				LastBlock: chunk.BVs[len(chunk.BVs)-1].LastBlock(),
				ID:        chunk.BVs[len(chunk.BVs)-1].EventLocator.ID(),
			}
		}

		_ = pm.bvLeecher.NotifyChunkReceived(chunk.SessionID, last, chunk.Done)

	case msg.Code == RequestBRsStream:
		var request brstream.Request
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if request.Limit.Num > hardLimitItems-1 {
			return errResp(ErrMsgTooLarge, "%v", msg)
		}
		if request.Limit.Size > protocolMaxMsgSize*2/3 {
			return errResp(ErrMsgTooLarge, "%v", msg)
		}

		pid := p.id
		_, peerErr := pm.brSeeder.NotifyRequestReceived(brstreamseeder.Peer{
			ID:        pid,
			SendChunk: p.SendBRsStream,
			Misbehaviour: func(err error) {
				pm.peerMisbehaviour(pid, err)
			},
		}, request)
		if peerErr != nil {
			return peerErr
		}

	case msg.Code == BRsStreamResponse:
		msgSize := uint64(msg.Size)
		var chunk brsChunk
		if err := msg.Decode(&chunk); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if err := checkLenLimits(len(chunk.BRs)+1, chunk); err != nil {
			return err
		}

		var last idx.Block
		if len(chunk.BRs) != 0 {
			_ = pm.brProcessor.Enqueue(p.id, chunk.BRs, msgSize, nil)
			last = chunk.BRs[len(chunk.BRs)-1].Idx
		}

		_ = pm.brLeecher.NotifyChunkReceived(chunk.SessionID, last, chunk.Done)

	case msg.Code == RequestEPsStream:
		var request epstream.Request
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if request.Limit.Num > hardLimitItems-1 {
			return errResp(ErrMsgTooLarge, "%v", msg)
		}
		if request.Limit.Size > protocolMaxMsgSize*2/3 {
			return errResp(ErrMsgTooLarge, "%v", msg)
		}

		pid := p.id
		_, peerErr := pm.epSeeder.NotifyRequestReceived(epstreamseeder.Peer{
			ID:        pid,
			SendChunk: p.SendEPsStream,
			Misbehaviour: func(err error) {
				pm.peerMisbehaviour(pid, err)
			},
		}, request)
		if peerErr != nil {
			return peerErr
		}

	case msg.Code == EPsStreamResponse:
		msgSize := uint64(msg.Size)
		var chunk epsChunk
		if err := msg.Decode(&chunk); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if err := checkLenLimits(len(chunk.EPs)+1, chunk); err != nil {
			return err
		}

		var last idx.Epoch
		if len(chunk.EPs) != 0 {
			_ = pm.epProcessor.Enqueue(p.id, chunk.EPs, msgSize, nil)
			last = chunk.EPs[len(chunk.EPs)-1].Record.Idx
		}

		_ = pm.epLeecher.NotifyChunkReceived(chunk.SessionID, last, chunk.Done)

	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}

func (pm *ProtocolManager) decideBroadcastAggressiveness(size int, passed time.Duration, peersNum int) int {
	percents := 100
	maxPercents := 1000000 * percents
	latencyVsThroughputTradeoff := maxPercents
	cfg := pm.config.Protocol
	if cfg.ThroughputImportance != 0 {
		latencyVsThroughputTradeoff = (cfg.LatencyImportance * percents) / cfg.ThroughputImportance
	}

	broadcastCost := passed * time.Duration(128+size) / 128
	broadcastAllCostTarget := time.Duration(latencyVsThroughputTradeoff) * (700 * time.Millisecond) / time.Duration(percents)
	broadcastSqrtCostTarget := broadcastAllCostTarget * 10

	fullRecipients := 0
	if latencyVsThroughputTradeoff >= maxPercents {
		// edge case
		fullRecipients = peersNum
	} else if latencyVsThroughputTradeoff <= 0 {
		// edge case
		fullRecipients = 0
	} else if broadcastCost <= broadcastAllCostTarget {
		// if event is small or was created recently, always send to everyone full event
		fullRecipients = peersNum
	} else if broadcastCost <= broadcastSqrtCostTarget || passed == 0 {
		// if event is big but was created recently, send full event to subset of peers
		fullRecipients = int(math.Sqrt(float64(peersNum)))
		if fullRecipients < 4 {
			fullRecipients = 4
		}
	}
	if fullRecipients > peersNum {
		fullRecipients = peersNum
	}
	return fullRecipients
}

// BroadcastEvent will either propagate a event to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
func (pm *ProtocolManager) BroadcastEvent(event *inter.EventPayload, passed time.Duration) int {
	if passed < 0 {
		passed = 0
	}
	id := event.ID()
	peers := pm.peers.PeersWithoutEvent(id)
	if len(peers) == 0 {
		log.Trace("Event is already known to all peers", "hash", id)
		return 0
	}

	fullRecipients := pm.decideBroadcastAggressiveness(event.Size(), passed, len(peers))

	// Broadcast of full event to a subset of peers
	fullBroadcast := peers[:fullRecipients]
	hashBroadcast := peers[fullRecipients:]
	for _, peer := range fullBroadcast {
		peer.AsyncSendEvents(inter.EventPayloads{event}, peer.queue)
	}
	// Broadcast of event hash to the rest peers
	for _, peer := range hashBroadcast {
		peer.AsyncSendEventIDs(hash.Events{event.ID()}, peer.queue)
	}
	log.Trace("Broadcast event", "hash", id, "fullRecipients", len(fullBroadcast), "hashRecipients", len(hashBroadcast))
	return len(peers)
}

// BroadcastTxs will propagate a batch of transactions to all peers which are not known to
// already have the given transaction.
func (pm *ProtocolManager) BroadcastTxs(txs types.Transactions) {
	var txset = make(map[*peer]types.Transactions)

	// Broadcast transactions to a batch of peers not knowing about it
	totalSize := common.StorageSize(0)
	for _, tx := range txs {
		peers := pm.peers.PeersWithoutTx(tx.Hash())
		for _, peer := range peers {
			txset[peer] = append(txset[peer], tx)
		}
		totalSize += tx.Size()
		log.Trace("Broadcast transaction", "hash", tx.Hash(), "recipients", len(peers))
	}
	fullRecipients := pm.decideBroadcastAggressiveness(int(totalSize), time.Second, len(txset))
	i := 0
	for peer, txs := range txset {
		SplitTransactions(txs, func(batch types.Transactions) {
			if i < fullRecipients {
				peer.AsyncSendTransactions(batch, peer.queue)
			} else {
				txids := make([]common.Hash, batch.Len())
				for i, tx := range batch {
					txids[i] = tx.Hash()
				}
				peer.AsyncSendTransactionHashes(txids, peer.queue)
			}
		})
		i++
	}
}

// Mined broadcast loop
func (pm *ProtocolManager) emittedBroadcastLoop() {
	defer pm.loopsWg.Done()
	for {
		select {
		case emitted := <-pm.emittedEventsCh:
			pm.BroadcastEvent(emitted, 0)
		// Err() channel will be closed when unsubscribing.
		case <-pm.emittedEventsSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) broadcastProgress() {
	progress := pm.myProgress()
	for _, peer := range pm.peers.List() {
		peer.AsyncSendProgress(progress, peer.queue)
	}
}

// Progress broadcast loop
func (pm *ProtocolManager) progressBroadcastLoop() {
	ticker := time.NewTicker(pm.config.Protocol.ProgressBroadcastPeriod)
	defer ticker.Stop()
	defer pm.loopsWg.Done()
	// automatically stops if unsubscribe
	for {
		select {
		case <-ticker.C:
			pm.broadcastProgress()
		case <-pm.quitProgressBradcast:
			return
		}
	}
}

func (pm *ProtocolManager) onNewEpochLoop() {
	defer pm.loopsWg.Done()
	for {
		select {
		case myEpoch := <-pm.newEpochsCh:
			pm.dagProcessor.Clear()
			if atomic.LoadUint32(&pm.synced) == 0 {
				synced := false
				for _, peer := range pm.peers.List() {
					if peer.progress.Epoch == myEpoch {
						synced = true
					}
				}
				// Mark initial sync done on any peer which has the same epoch
				if synced {
					atomic.StoreUint32(&pm.synced, 1)
				}
			}
			pm.dagLeecher.OnNewEpoch(myEpoch)
		// Err() channel will be closed when unsubscribing.
		case <-pm.newEpochsSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) txBroadcastLoop() {
	ticker := time.NewTicker(pm.config.Protocol.RandomTxHashesSendPeriod)
	defer ticker.Stop()
	defer pm.loopsWg.Done()
	for {
		select {
		case notify := <-pm.txsCh:
			pm.BroadcastTxs(notify.Txs)

		// Err() channel will be closed when unsubscribing.
		case <-pm.txsSub.Err():
			return

		case <-ticker.C:
			if atomic.LoadUint32(&pm.synced) == 0 {
				continue
			}
			peers := pm.peers.List()
			if len(peers) == 0 {
				continue
			}
			randPeer := peers[rand.Intn(len(peers))]
			pm.syncTransactions(randPeer, pm.txpool.SampleHashes(pm.config.Protocol.MaxRandomTxHashesSend))
		}
	}
}

// NodeInfo represents a short summary of the sub-protocol metadata
// known about the host peer.
type NodeInfo struct {
	Network     uint64      `json:"network"` // network ID
	Genesis     common.Hash `json:"genesis"` // SHA3 hash of the host's genesis object
	Epoch       idx.Epoch   `json:"epoch"`
	NumOfBlocks idx.Block   `json:"blocks"`
	//Config  *params.ChainConfig `json:"config"`  // Chain configuration for the fork rules
}

// NodeInfo retrieves some protocol metadata about the running host node.
func (pm *ProtocolManager) NodeInfo() *NodeInfo {
	numOfBlocks := pm.store.GetLatestBlockIndex()
	return &NodeInfo{
		Network:     pm.NetworkID,
		Genesis:     common.Hash(*pm.store.GetGenesisHash()),
		Epoch:       pm.store.GetEpoch(),
		NumOfBlocks: numOfBlocks,
	}
}

func getSemaphoreWarningFn(name string) func(dag.Metric, dag.Metric, dag.Metric) {
	return func(received dag.Metric, processing dag.Metric, releasing dag.Metric) {
		log.Warn(fmt.Sprintf("%s semaphore inconsistency", name),
			"receivedNum", received.Num, "receivedSize", received.Size,
			"processingNum", processing.Num, "processingSize", processing.Size,
			"releasingNum", releasing.Num, "releasingSize", releasing.Size)
	}
}
