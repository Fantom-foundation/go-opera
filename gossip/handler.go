package gossip

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
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
	"github.com/ethereum/go-ethereum/p2p/discover/discfilter"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"

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
	"github.com/Fantom-foundation/go-opera/gossip/protocols/snap/snapstream/snapleecher"
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

	sameMsgDisconnectThreshold = 10
	maxMsgFrequency            = 1 * time.Second
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
	Event            func(*inter.EventPayload) error
	SwitchEpochTo    func(idx.Epoch) error
	PauseEvmSnapshot func()
	BVs              func(inter.LlrSignedBlockVotes) error
	BR               func(ibr.LlrIdxFullBlockRecord) error
	EV               func(inter.LlrSignedEpochVote) error
	ER               func(ier.LlrIdxFullEpochRecord) error
}

// handlerConfig is the collection of initialization parameters to create a full
// node network handler.
type handlerConfig struct {
	config   Config
	notifier dagNotifier
	txpool   TxPool
	engineMu sync.Locker
	checkers *eventcheck.Checkers
	s        *Store
	process  processCallback
}

type snapsyncEpochUpd struct {
	epoch idx.Epoch
	root  common.Hash
}

type snapsyncCancelCmd struct {
	done chan struct{}
}

type snapsyncStateUpd struct {
	snapsyncEpochUpd  *snapsyncEpochUpd
	snapsyncCancelCmd *snapsyncCancelCmd
}

type snapsyncState struct {
	epoch     idx.Epoch
	cancel    func() error
	updatesCh chan snapsyncStateUpd
	quit      chan struct{}
}

type msgDOSwatch struct {
	timestamp []time.Time
	count     int
}

type handler struct {
	NetworkID uint64
	config    Config

	syncStatus syncStatus

	txpool   TxPool
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
	txsyncCh chan *txsync
	quitSync chan struct{}

	// snapsync fields
	chain       *ethBlockChain
	snapLeecher *snapleecher.Leecher
	snapState   snapsyncState

	// wait group is used for graceful shutdowns during downloading
	// and processing
	loopsWg sync.WaitGroup
	wg      sync.WaitGroup
	peerWG  sync.WaitGroup
	started sync.WaitGroup

	msgCache map[uint64]map[string]*msgDOSwatch

	logger.Instance
}

// newHandler returns a new Fantom sub protocol manager. The Fantom sub protocol manages peers capable
// with the Fantom network.
func newHandler(
	c handlerConfig,
) (
	*handler,
	error,
) {
	// Create the protocol manager with the base fields
	h := &handler{
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
		txsyncCh:             make(chan *txsync),
		quitSync:             make(chan struct{}),
		quitProgressBradcast: make(chan struct{}),
		msgCache:             make(map[uint64]map[string]*msgDOSwatch),

		snapState: snapsyncState{
			updatesCh: make(chan snapsyncStateUpd, 128),
			quit:      make(chan struct{}),
		},

		Instance: logger.New("PM"),
	}
	h.started.Add(1)

	// TODO: configure it
	var (
		configBloomCache uint64 = 0 // Megabytes to alloc for fast sync bloom
	)

	var err error
	h.chain, err = newEthBlockChain(c.s)
	if err != nil {
		return nil, err
	}

	stateDb := h.store.EvmStore().EvmDb
	var stateBloom *trie.SyncBloom
	if false {
		// NOTE: Construct the downloader (long sync) and its backing state bloom if fast
		// sync is requested. The downloader is responsible for deallocating the state
		// bloom when it's done.
		// Note: we don't enable it if snap-sync is performed, since it's very heavy
		// and the heal-portion of the snap sync is much lighter than fast. What we particularly
		// want to avoid, is a 90%-finished (but restarted) snap-sync to begin
		// indexing the entire trie
		stateBloom = trie.NewSyncBloom(configBloomCache, stateDb)
	}
	h.snapLeecher = snapleecher.New(stateDb, stateBloom, h.removePeer)

	h.dagFetcher = itemsfetcher.New(h.config.Protocol.DagFetcher, itemsfetcher.Callback{
		OnlyInterested: func(ids []interface{}) []interface{} {
			return h.onlyInterestedEventsI(ids)
		},
		Suspend: func() bool {
			return false
		},
	})
	h.txFetcher = itemsfetcher.New(h.config.Protocol.TxFetcher, itemsfetcher.Callback{
		OnlyInterested: func(txids []interface{}) []interface{} {
			return txidsToInterfaces(h.txpool.OnlyNotExisting(interfacesToTxids(txids)))
		},
		Suspend: func() bool {
			return false
		},
	})

	h.dagProcessor = h.makeDagProcessor(c.checkers)
	h.dagLeecher = dagstreamleecher.New(h.store.GetEpoch(), h.store.GetHighestLamport() == 0, h.config.Protocol.DagStreamLeecher, dagstreamleecher.Callbacks{
		IsProcessed: h.store.HasEvent,
		RequestChunk: func(peer string, r dagstream.Request) error {
			p := h.peers.Peer(peer)
			if p == nil {
				return errNotRegistered
			}
			return p.RequestEventsStream(r)
		},
		Suspend: func(_ string) bool {
			return h.dagFetcher.Overloaded() || h.dagProcessor.Overloaded()
		},
		PeerEpoch: func(peer string) idx.Epoch {
			p := h.peers.Peer(peer)
			if p == nil || p.Useless() {
				return 0
			}
			return p.progress.Epoch
		},
	})
	h.dagSeeder = dagstreamseeder.New(h.config.Protocol.DagStreamSeeder, dagstreamseeder.Callbacks{
		ForEachEvent: c.s.ForEachEventRLP,
	})

	h.bvProcessor = h.makeBvProcessor(c.checkers)
	h.bvLeecher = bvstreamleecher.New(h.config.Protocol.BvStreamLeecher, bvstreamleecher.Callbacks{
		LowestBlockToDecide: func() (idx.Epoch, idx.Block) {
			llrs := h.store.GetLlrState()
			epoch := h.store.FindBlockEpoch(llrs.LowestBlockToDecide)
			return epoch, llrs.LowestBlockToDecide
		},
		MaxEpochToDecide: func() idx.Epoch {
			if !h.syncStatus.RequestLLR() {
				return 0
			}
			return h.store.GetLlrState().LowestEpochToFill
		},
		IsProcessed: h.store.HasBlockVotes,
		RequestChunk: func(peer string, r bvstream.Request) error {
			p := h.peers.Peer(peer)
			if p == nil {
				return errNotRegistered
			}
			return p.RequestBVsStream(r)
		},
		Suspend: func(_ string) bool {
			return h.bvProcessor.Overloaded()
		},
		PeerBlock: func(peer string) idx.Block {
			p := h.peers.Peer(peer)
			if p == nil || p.Useless() {
				return 0
			}
			return p.progress.LastBlockIdx
		},
	})
	h.bvSeeder = bvstreamseeder.New(h.config.Protocol.BvStreamSeeder, bvstreamseeder.Callbacks{
		Iterate: h.store.IterateOverlappingBlockVotesRLP,
	})

	h.brProcessor = h.makeBrProcessor()
	h.brLeecher = brstreamleecher.New(h.config.Protocol.BrStreamLeecher, brstreamleecher.Callbacks{
		LowestBlockToFill: func() idx.Block {
			return h.store.GetLlrState().LowestBlockToFill
		},
		MaxBlockToFill: func() idx.Block {
			if !h.syncStatus.RequestLLR() {
				return 0
			}
			// rough estimation for the max fill-able block
			llrs := h.store.GetLlrState()
			start := llrs.LowestBlockToFill
			end := llrs.LowestBlockToDecide
			if end > start+100 && h.store.HasBlock(start+100) {
				return start + 100
			}
			return end
		},
		IsProcessed: h.store.HasBlock,
		RequestChunk: func(peer string, r brstream.Request) error {
			p := h.peers.Peer(peer)
			if p == nil {
				return errNotRegistered
			}
			return p.RequestBRsStream(r)
		},
		Suspend: func(_ string) bool {
			return h.brProcessor.Overloaded()
		},
		PeerBlock: func(peer string) idx.Block {
			p := h.peers.Peer(peer)
			if p == nil || p.Useless() {
				return 0
			}
			return p.progress.LastBlockIdx
		},
	})
	h.brSeeder = brstreamseeder.New(h.config.Protocol.BrStreamSeeder, brstreamseeder.Callbacks{
		Iterate: h.store.IterateFullBlockRecordsRLP,
	})

	h.epProcessor = h.makeEpProcessor(h.checkers)
	h.epLeecher = epstreamleecher.New(h.config.Protocol.EpStreamLeecher, epstreamleecher.Callbacks{
		LowestEpochToFetch: func() idx.Epoch {
			llrs := h.store.GetLlrState()
			if llrs.LowestEpochToFill < llrs.LowestEpochToDecide {
				return llrs.LowestEpochToFill
			}
			return llrs.LowestEpochToDecide
		},
		MaxEpochToFetch: func() idx.Epoch {
			if !h.syncStatus.RequestLLR() {
				return 0
			}
			return h.store.GetLlrState().LowestEpochToDecide + 10000
		},
		IsProcessed: h.store.HasHistoryBlockEpochState,
		RequestChunk: func(peer string, r epstream.Request) error {
			p := h.peers.Peer(peer)
			if p == nil {
				return errNotRegistered
			}
			return p.RequestEPsStream(r)
		},
		Suspend: func(_ string) bool {
			return h.epProcessor.Overloaded()
		},
		PeerEpoch: func(peer string) idx.Epoch {
			p := h.peers.Peer(peer)
			if p == nil || p.Useless() {
				return 0
			}
			return p.progress.Epoch
		},
	})
	h.epSeeder = epstreamseeder.New(h.config.Protocol.EpStreamSeeder, epstreamseeder.Callbacks{
		Iterate: h.store.IterateEpochPacksRLP,
	})

	return h, nil
}

func (h *handler) peerMisbehaviour(peer string, err error) bool {
	if eventcheck.IsBan(err) {
		log.Warn("Dropping peer due to a misbehaviour", "peer", peer, "err", err)
		h.removePeer(peer)
		return true
	}
	return false
}

func (h *handler) makeDagProcessor(checkers *eventcheck.Checkers) *dagprocessor.Processor {
	// checkers
	lightCheck := func(e dag.Event) error {
		if h.store.GetEpoch() != e.ID().Epoch() {
			return epochcheck.ErrNotRelevant
		}
		if h.dagProcessor.IsBuffered(e.ID()) {
			return eventcheck.ErrDuplicateEvent
		}
		if h.store.HasEvent(e.ID()) {
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
	newProcessor := dagprocessor.New(datasemaphore.New(h.config.Protocol.EventsSemaphoreLimit, getSemaphoreWarningFn("DAG events")), h.config.Protocol.DagProcessor, dagprocessor.Callback{
		// DAG callbacks
		Event: dagprocessor.EventCallback{
			Process: func(_e dag.Event) error {
				e := _e.(*inter.EventPayload)
				preStart := time.Now()
				h.engineMu.Lock()
				defer h.engineMu.Unlock()

				err := h.process.Event(e)
				if err != nil {
					return err
				}

				// event is connected, announce it
				passedSinceEvent := preStart.Sub(e.CreationTime().Time())
				h.BroadcastEvent(e, passedSinceEvent)

				return nil
			},
			Released: func(e dag.Event, peer string, err error) {
				if eventcheck.IsBan(err) {
					log.Warn("Incoming event rejected", "event", e.ID().String(), "creator", e.Creator(), "err", err)
					h.removePeer(peer)
				}
			},

			Exists: func(id hash.Event) bool {
				return h.store.HasEvent(id)
			},

			Get: func(id hash.Event) dag.Event {
				e := h.store.GetEventPayload(id)
				if e == nil {
					return nil
				}
				return e
			},

			CheckParents:    bufferedCheck,
			CheckParentless: parentlessChecker.Enqueue,
		},
		HighestLamport: h.store.GetHighestLamport,
	})

	return newProcessor
}

func (h *handler) makeBvProcessor(checkers *eventcheck.Checkers) *bvprocessor.Processor {
	// checkers
	lightCheck := func(bvs inter.LlrSignedBlockVotes) error {
		if h.store.HasBlockVotes(bvs.Val.Epoch, bvs.Val.LastBlock(), bvs.Signed.Locator.ID()) {
			return eventcheck.ErrAlreadyProcessedBVs
		}
		return checkers.Basiccheck.ValidateBVs(bvs)
	}
	allChecker := bvallcheck.Checker{
		HeavyCheck: &heavycheck.BVsOnly{Checker: checkers.Heavycheck},
		LightCheck: lightCheck,
	}
	return bvprocessor.New(datasemaphore.New(h.config.Protocol.BVsSemaphoreLimit, getSemaphoreWarningFn("BVs")), h.config.Protocol.BvProcessor, bvprocessor.Callback{
		// DAG callbacks
		Item: bvprocessor.ItemCallback{
			Process: h.process.BVs,
			Released: func(bvs inter.LlrSignedBlockVotes, peer string, err error) {
				if eventcheck.IsBan(err) {
					log.Warn("Incoming BVs rejected", "BVs", bvs.Signed.Locator.ID(), "creator", bvs.Signed.Locator.Creator, "err", err)
					h.removePeer(peer)
				}
			},
			Check: allChecker.Enqueue,
		},
	})
}

func (h *handler) makeBrProcessor() *brprocessor.Processor {
	// checkers
	return brprocessor.New(datasemaphore.New(h.config.Protocol.BVsSemaphoreLimit, getSemaphoreWarningFn("BR")), h.config.Protocol.BrProcessor, brprocessor.Callback{
		// DAG callbacks
		Item: brprocessor.ItemCallback{
			Process: h.process.BR,
			Released: func(br ibr.LlrIdxFullBlockRecord, peer string, err error) {
				if eventcheck.IsBan(err) {
					log.Warn("Incoming BR rejected", "block", br.Idx, "err", err)
					h.removePeer(peer)
				}
			},
		},
	})
}

func (h *handler) makeEpProcessor(checkers *eventcheck.Checkers) *epprocessor.Processor {
	// checkers
	lightCheck := func(ev inter.LlrSignedEpochVote) error {
		if h.store.HasEpochVote(ev.Val.Epoch, ev.Signed.Locator.ID()) {
			return eventcheck.ErrAlreadyProcessedEV
		}
		return checkers.Basiccheck.ValidateEV(ev)
	}
	allChecker := evallcheck.Checker{
		HeavyCheck: &heavycheck.EVOnly{Checker: checkers.Heavycheck},
		LightCheck: lightCheck,
	}
	// checkers
	return epprocessor.New(datasemaphore.New(h.config.Protocol.BVsSemaphoreLimit, getSemaphoreWarningFn("BR")), h.config.Protocol.EpProcessor, epprocessor.Callback{
		// DAG callbacks
		Item: epprocessor.ItemCallback{
			ProcessEV: h.process.EV,
			ProcessER: h.process.ER,
			ReleasedEV: func(ev inter.LlrSignedEpochVote, peer string, err error) {
				if eventcheck.IsBan(err) {
					log.Warn("Incoming EV rejected", "event", ev.Signed.Locator.ID(), "creator", ev.Signed.Locator.Creator, "err", err)
					h.removePeer(peer)
				}
			},
			ReleasedER: func(er ier.LlrIdxFullEpochRecord, peer string, err error) {
				if eventcheck.IsBan(err) {
					log.Warn("Incoming ER rejected", "epoch", er.Idx, "err", err)
					h.removePeer(peer)
				}
			},
			CheckEV: allChecker.Enqueue,
		},
	})
}

func (h *handler) isEventInterested(id hash.Event, epoch idx.Epoch) bool {
	if id.Epoch() != epoch {
		return false
	}

	if h.dagProcessor.IsBuffered(id) || h.store.HasEvent(id) {
		return false
	}
	return true
}

func (h *handler) onlyInterestedEventsI(ids []interface{}) []interface{} {
	if len(ids) == 0 {
		return ids
	}
	epoch := h.store.GetEpoch()
	interested := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		if h.isEventInterested(id.(hash.Event), epoch) {
			interested = append(interested, id)
		}
	}
	return interested
}

func (h *handler) removePeer(id string) {
	peer := h.peers.Peer(id)
	if peer != nil {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
	}
}

func (h *handler) unregisterPeer(id string) {
	// Short circuit if the peer was already removed
	peer := h.peers.Peer(id)
	if peer == nil {
		return
	}
	log.Debug("Removing peer", "peer", id)

	// Unregister the peer from the leecher's and seeder's and peer sets
	_ = h.epLeecher.UnregisterPeer(id)
	_ = h.epSeeder.UnregisterPeer(id)
	_ = h.dagLeecher.UnregisterPeer(id)
	_ = h.dagSeeder.UnregisterPeer(id)
	_ = h.brLeecher.UnregisterPeer(id)
	_ = h.brSeeder.UnregisterPeer(id)
	_ = h.bvLeecher.UnregisterPeer(id)
	_ = h.bvSeeder.UnregisterPeer(id)
	// Remove the `snap` extension if it exists
	if peer.snapExt != nil {
		_ = h.snapLeecher.SnapSyncer.Unregister(id)
	}
	if err := h.peers.UnregisterPeer(id); err != nil {
		log.Error("Peer removal failed", "peer", id, "err", err)
	}
}

func (h *handler) Start(maxPeers int) {
	h.snapsyncStageTick()

	h.maxPeers = maxPeers

	// broadcast transactions
	h.txsCh = make(chan evmcore.NewTxsNotify, txChanSize)
	h.txsSub = h.txpool.SubscribeNewTxsNotify(h.txsCh)

	h.loopsWg.Add(1)
	go h.txBroadcastLoop()

	if h.notifier != nil {
		// broadcast mined events
		h.emittedEventsCh = make(chan *inter.EventPayload, 4)
		h.emittedEventsSub = h.notifier.SubscribeNewEmitted(h.emittedEventsCh)
		// epoch changes
		h.newEpochsCh = make(chan idx.Epoch, 4)
		h.newEpochsSub = h.notifier.SubscribeNewEpoch(h.newEpochsCh)

		h.loopsWg.Add(3)
		go h.emittedBroadcastLoop()
		go h.progressBroadcastLoop()
		go h.onNewEpochLoop()
	}

	// start sync handlers
	go h.txsyncLoop()
	h.loopsWg.Add(2)
	go h.snapsyncStateLoop()
	go h.snapsyncStageLoop()
	h.dagFetcher.Start()
	h.txFetcher.Start()
	h.checkers.Heavycheck.Start()

	h.epProcessor.Start()
	h.epSeeder.Start()
	h.epLeecher.Start()

	h.dagProcessor.Start()
	h.dagSeeder.Start()
	h.dagLeecher.Start()

	h.bvProcessor.Start()
	h.bvSeeder.Start()
	h.bvLeecher.Start()

	h.brProcessor.Start()
	h.brSeeder.Start()
	h.brLeecher.Start()
	h.started.Done()
}

func (h *handler) Stop() {
	log.Info("Stopping Fantom protocol")

	h.brLeecher.Stop()
	h.brSeeder.Stop()
	h.brProcessor.Stop()

	h.bvLeecher.Stop()
	h.bvSeeder.Stop()
	h.bvProcessor.Stop()

	h.dagLeecher.Stop()
	h.dagSeeder.Stop()
	h.dagProcessor.Stop()

	h.epLeecher.Stop()
	h.epSeeder.Stop()
	h.epProcessor.Stop()

	h.checkers.Heavycheck.Stop()
	h.txFetcher.Stop()
	h.dagFetcher.Stop()

	close(h.quitProgressBradcast)
	close(h.snapState.quit)
	h.txsSub.Unsubscribe() // quits txBroadcastLoop
	if h.notifier != nil {
		h.emittedEventsSub.Unsubscribe() // quits eventBroadcastLoop
		h.newEpochsSub.Unsubscribe()     // quits onNewEpochLoop
	}

	// Wait for the subscription loops to come down.
	h.loopsWg.Wait()

	h.msgSemaphore.Terminate()
	// Quit the sync loop.
	// After this send has completed, no new peers will be accepted.
	close(h.quitSync)

	// Disconnect existing sessions.
	// This also closes the gate for any new registrations on the peer set.
	// sessions which are already established but not added to h.peers yet
	// will exit when they try to register.
	h.peers.Close()

	// Wait for all peer handler goroutines to come down.
	h.wg.Wait()
	h.peerWG.Wait()

	log.Info("Fantom protocol stopped")
}

func (h *handler) myProgress() PeerProgress {
	bs := h.store.GetBlockState()
	epoch := h.store.GetEpoch()
	return PeerProgress{
		Epoch:            epoch,
		LastBlockIdx:     bs.LastBlock.Idx,
		LastBlockAtropos: bs.LastBlock.Atropos,
	}
}

func (h *handler) highestPeerProgress() PeerProgress {
	peers := h.peers.List()
	max := h.myProgress()
	for _, peer := range peers {
		if max.LastBlockIdx < peer.progress.LastBlockIdx {
			max = peer.progress
		}
	}
	return max
}

// handle is the callback invoked to manage the life cycle of a peer. When
// this function terminates, the peer is disconnected.
func (h *handler) handle(p *peer) error {
	// If the peer has a `snap` extension, wait for it to connect so we can have
	// a uniform initialization/teardown mechanism
	snap, err := h.peers.WaitSnapExtension(p)
	if err != nil {
		p.Log().Error("Snapshot extension barrier failed", "err", err)
		return err
	}
	useless := discfilter.Banned(p.Node().ID(), p.Node().Record())
	if !useless && (!eligibleForSnap(p.Peer) || !strings.Contains(strings.ToLower(p.Name()), "opera")) {
		useless = true
		discfilter.Ban(p.ID())
	}
	if !p.Peer.Info().Network.Trusted && useless && h.peers.UselessNum() >= h.maxPeers/10 {
		// don't allow more than 10% of useless peers
		return p2p.DiscTooManyPeers
	}
	if !p.Peer.Info().Network.Trusted && useless {
		if h.peers.UselessNum() >= h.maxPeers/10 {
			// don't allow more than 10% of useless peers
			return p2p.DiscTooManyPeers
		}
		p.SetUseless()
	}

	h.peerWG.Add(1)
	defer h.peerWG.Done()

	// Execute the handshake
	var (
		genesis    = *h.store.GetGenesisID()
		myProgress = h.myProgress()
	)
	if err := p.Handshake(h.NetworkID, myProgress, common.Hash(genesis)); err != nil {
		p.Log().Debug("Handshake failed", "err", err)
		if !useless {
			discfilter.Ban(p.ID())
		}
		return err
	}

	// Ignore maxPeers if this is a trusted peer
	if h.peers.Len() >= h.maxPeers && !p.Peer.Info().Network.Trusted {
		return p2p.DiscTooManyPeers
	}
	p.Log().Debug("Peer connected", "name", p.Name())

	// Register the peer locally
	if err := h.peers.RegisterPeer(p, snap); err != nil {
		p.Log().Warn("Peer registration failed", "err", err)
		return err
	}
	if err := h.dagLeecher.RegisterPeer(p.id); err != nil {
		p.Log().Warn("Leecher peer registration failed", "err", err)
		return err
	}
	if p.RunningCap(ProtocolName, []uint{FTM63}) {
		if err := h.epLeecher.RegisterPeer(p.id); err != nil {
			p.Log().Warn("Leecher peer registration failed", "err", err)
			return err
		}
		if err := h.bvLeecher.RegisterPeer(p.id); err != nil {
			p.Log().Warn("Leecher peer registration failed", "err", err)
			return err
		}
		if err := h.brLeecher.RegisterPeer(p.id); err != nil {
			p.Log().Warn("Leecher peer registration failed", "err", err)
			return err
		}
	}
	if snap != nil {
		if err := h.snapLeecher.SnapSyncer.Register(snap); err != nil {
			p.Log().Error("Failed to register peer in snap syncer", "err", err)
			return err
		}
	}
	defer h.unregisterPeer(p.id)

	// Propagate existing transactions. new transactions appearing
	// after this will be sent via broadcasts.
	h.syncTransactions(p, h.txpool.SampleHashes(h.config.Protocol.MaxInitialTxHashesSend))

	// Handle incoming messages until the connection is torn down
	for {
		if err := h.handleMsg(p); err != nil {
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

func (h *handler) handleTxHashes(p *peer, announces []common.Hash) {
	// Mark the hashes as present at the remote node
	for _, id := range announces {
		p.MarkTransaction(id)
	}
	// Schedule all the unknown hashes for retrieval
	requestTransactions := func(ids []interface{}) error {
		return p.RequestTransactions(interfacesToTxids(ids))
	}
	_ = h.txFetcher.NotifyAnnounces(p.id, txidsToInterfaces(announces), time.Now(), requestTransactions)
}

func (h *handler) handleTxs(p *peer, txs types.Transactions) {
	// Mark the hashes as present at the remote node
	for _, tx := range txs {
		p.MarkTransaction(tx.Hash())
	}
	h.txpool.AddRemotes(txs)
}

func (h *handler) handleEventHashes(p *peer, announces hash.Events) {
	// Mark the hashes as present at the remote node
	for _, id := range announces {
		p.MarkEvent(id)
	}
	// filter too high IDs
	notTooHigh := make(hash.Events, 0, len(announces))
	sessionCfg := h.config.Protocol.DagStreamLeecher.Session
	for _, id := range announces {
		maxLamport := h.store.GetHighestLamport() + idx.Lamport(sessionCfg.DefaultChunkItemsNum+1)*idx.Lamport(sessionCfg.ParallelChunksDownload)
		if id.Lamport() <= maxLamport {
			notTooHigh = append(notTooHigh, id)
		}
	}
	if len(announces) != len(notTooHigh) {
		h.dagLeecher.ForceSyncing()
	}
	if len(notTooHigh) == 0 {
		return
	}
	// Schedule all the unknown hashes for retrieval
	requestEvents := func(ids []interface{}) error {
		return p.RequestEvents(interfacesToEventIDs(ids))
	}
	_ = h.dagFetcher.NotifyAnnounces(p.id, eventIDsToInterfaces(notTooHigh), time.Now(), requestEvents)
}

func (h *handler) handleEvents(p *peer, events dag.Events, ordered bool) {
	// Mark the hashes as present at the remote node
	for _, e := range events {
		p.MarkEvent(e.ID())
	}
	// filter too high events
	notTooHigh := make(dag.Events, 0, len(events))
	sessionCfg := h.config.Protocol.DagStreamLeecher.Session
	now := time.Now()
	for _, e := range events {
		maxLamport := h.store.GetHighestLamport() + idx.Lamport(sessionCfg.DefaultChunkItemsNum+1)*idx.Lamport(sessionCfg.ParallelChunksDownload)
		if e.Lamport() <= maxLamport {
			notTooHigh = append(notTooHigh, e)
		}
		if now.Sub(e.(inter.EventI).CreationTime().Time()) < 10*time.Minute {
			h.syncStatus.MarkMaybeSynced()
		}
	}
	if len(events) != len(notTooHigh) {
		h.dagLeecher.ForceSyncing()
	}
	if len(notTooHigh) == 0 {
		return
	}
	// Schedule all the events for connection
	peer := *p
	requestEvents := func(ids []interface{}) error {
		return peer.RequestEvents(interfacesToEventIDs(ids))
	}
	notifyAnnounces := func(ids hash.Events) {
		_ = h.dagFetcher.NotifyAnnounces(peer.id, eventIDsToInterfaces(ids), now, requestEvents)
	}
	_ = h.dagProcessor.Enqueue(peer.id, notTooHigh, ordered, notifyAnnounces, nil)
}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (h *handler) handleMsg(p *peer) error {
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
	if !h.msgSemaphore.Acquire(eventsSizeEst, h.config.Protocol.MsgsSemaphoreTimeout) {
		h.Log.Warn("Failed to acquire semaphore for p2p message", "size", msg.Size, "peer", p.id)
		return nil
	}
	defer h.msgSemaphore.Release(eventsSizeEst)

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

	case msg.Code == EvmTxsMsg:
		// Transactions arrived, make sure we have a valid and fresh graph to handle them
		if !h.syncStatus.AcceptTxs() {
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
		_ = h.txFetcher.NotifyReceived(txids)
		h.handleTxs(p, txs)

	case msg.Code == NewEvmTxHashesMsg:
		// Transactions arrived, make sure we have a valid and fresh graph to handle them
		if !h.syncStatus.AcceptTxs() {
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
		h.handleTxHashes(p, txHashes)

	case msg.Code == GetEvmTxsMsg:
		var requests []common.Hash
		if err := msg.Decode(&requests); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		if err := checkLenLimits(len(requests), requests); err != nil {
			return err
		}

		txs := make(types.Transactions, 0, len(requests))
		for _, txid := range requests {
			tx := h.txpool.Get(txid)
			if tx == nil {
				continue
			}
			txs = append(txs, tx)
		}
		SplitTransactions(txs, func(batch types.Transactions) {
			p.EnqueueSendTransactions(batch, p.queue)
		})

	case msg.Code == EventsMsg:
		if !h.syncStatus.AcceptEvents() {
			break
		}

		var events inter.EventPayloads
		if err := msg.Decode(&events); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if err := checkLenLimits(len(events), events); err != nil {
			return err
		}
		_ = h.dagFetcher.NotifyReceived(eventIDsToInterfaces(events.IDs()))
		h.handleEvents(p, events.Bases(), events.Len() > 1)

	case msg.Code == NewEventIDsMsg:
		// Fresh events arrived, make sure we have a valid and fresh graph to handle them
		if !h.syncStatus.AcceptEvents() {
			break
		}
		var announces hash.Events
		if err := msg.Decode(&announces); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if err := checkLenLimits(len(announces), announces); err != nil {
			return err
		}
		h.handleEventHashes(p, announces)

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
			if raw := h.store.GetEventPayloadRLP(id); raw != nil {
				rawEvents = append(rawEvents, raw)
				ids = append(ids, id)
				size += len(raw)
			} else {
				h.Log.Debug("requested event not found", "hash", id)
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

		if h.msgCache[msg.Code] == nil {
			h.msgCache[msg.Code] = make(map[string]*msgDOSwatch)
		}
		id := msg.String()
		dosWatch := h.msgCache[msg.Code][id]
		if dosWatch == nil {
			dosWatch = new(msgDOSwatch)
		}
		dosWatch.timestamp = append(dosWatch.timestamp, time.Now())
		if len(dosWatch.timestamp) > sameMsgDisconnectThreshold {
			if time.Now().Sub(dosWatch.timestamp[0]) > maxMsgFrequency {
				return errors.New("too frequent")
			}
			dosWatch.timestamp = dosWatch.timestamp[:0]
		}

		pid := p.id
		_, peerErr := h.dagSeeder.NotifyRequestReceived(dagstreamseeder.Peer{
			ID:        pid,
			SendChunk: p.SendEventsStream,
			Misbehaviour: func(err error) {
				h.peerMisbehaviour(pid, err)
			},
		}, request)
		if peerErr != nil {
			return peerErr
		}

	case msg.Code == EventsStreamResponse:
		if !h.syncStatus.AcceptEvents() {
			break
		}

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
			h.handleEventHashes(p, chunk.IDs)
			last = chunk.IDs[len(chunk.IDs)-1]
		}
		if len(chunk.Events) != 0 {
			h.handleEvents(p, chunk.Events.Bases(), true)
			last = chunk.Events[len(chunk.Events)-1].ID()
		}

		_ = h.dagLeecher.NotifyChunkReceived(chunk.SessionID, last, chunk.Done)

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
		_, peerErr := h.bvSeeder.NotifyRequestReceived(bvstreamseeder.Peer{
			ID:        pid,
			SendChunk: p.SendBVsStream,
			Misbehaviour: func(err error) {
				h.peerMisbehaviour(pid, err)
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
			_ = h.bvProcessor.Enqueue(p.id, chunk.BVs, nil)
			last = bvstreamleecher.BVsID{
				Epoch:     chunk.BVs[len(chunk.BVs)-1].Val.Epoch,
				LastBlock: chunk.BVs[len(chunk.BVs)-1].Val.LastBlock(),
				ID:        chunk.BVs[len(chunk.BVs)-1].Signed.Locator.ID(),
			}
		}

		_ = h.bvLeecher.NotifyChunkReceived(chunk.SessionID, last, chunk.Done)

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
		_, peerErr := h.brSeeder.NotifyRequestReceived(brstreamseeder.Peer{
			ID:        pid,
			SendChunk: p.SendBRsStream,
			Misbehaviour: func(err error) {
				h.peerMisbehaviour(pid, err)
			},
		}, request)
		if peerErr != nil {
			return peerErr
		}

	case msg.Code == BRsStreamResponse:
		if !h.syncStatus.AcceptBlockRecords() {
			break
		}

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
			_ = h.brProcessor.Enqueue(p.id, chunk.BRs, msgSize, nil)
			last = chunk.BRs[len(chunk.BRs)-1].Idx
		}

		_ = h.brLeecher.NotifyChunkReceived(chunk.SessionID, last, chunk.Done)

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
		_, peerErr := h.epSeeder.NotifyRequestReceived(epstreamseeder.Peer{
			ID:        pid,
			SendChunk: p.SendEPsStream,
			Misbehaviour: func(err error) {
				h.peerMisbehaviour(pid, err)
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
			_ = h.epProcessor.Enqueue(p.id, chunk.EPs, msgSize, nil)
			last = chunk.EPs[len(chunk.EPs)-1].Record.Idx
		}

		_ = h.epLeecher.NotifyChunkReceived(chunk.SessionID, last, chunk.Done)

	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}

func (h *handler) decideBroadcastAggressiveness(size int, passed time.Duration, peersNum int) int {
	percents := 100
	maxPercents := 1000000 * percents
	latencyVsThroughputTradeoff := maxPercents
	cfg := h.config.Protocol
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
func (h *handler) BroadcastEvent(event *inter.EventPayload, passed time.Duration) int {
	if passed < 0 {
		passed = 0
	}
	id := event.ID()
	peers := h.peers.PeersWithoutEvent(id)
	if len(peers) == 0 {
		log.Trace("Event is already known to all peers", "hash", id)
		return 0
	}

	fullRecipients := h.decideBroadcastAggressiveness(event.Size(), passed, len(peers))

	// Exclude low quality peers from fullBroadcast
	var fullBroadcast = make([]*peer, 0, fullRecipients)
	var hashBroadcast = make([]*peer, 0, len(peers))
	for _, p := range peers {
		if !p.Useless() && len(fullBroadcast) < fullRecipients {
			fullBroadcast = append(fullBroadcast, p)
		} else {
			hashBroadcast = append(hashBroadcast, p)
		}
	}
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
func (h *handler) BroadcastTxs(txs types.Transactions) {
	var txset = make(map[*peer]types.Transactions)

	// Broadcast transactions to a batch of peers not knowing about it
	totalSize := common.StorageSize(0)
	for _, tx := range txs {
		peers := h.peers.PeersWithoutTx(tx.Hash())
		for _, peer := range peers {
			txset[peer] = append(txset[peer], tx)
		}
		totalSize += tx.Size()
		log.Trace("Broadcast transaction", "hash", tx.Hash(), "recipients", len(peers))
	}
	fullRecipients := h.decideBroadcastAggressiveness(int(totalSize), time.Second, len(txset))
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
func (h *handler) emittedBroadcastLoop() {
	defer h.loopsWg.Done()
	for {
		select {
		case emitted := <-h.emittedEventsCh:
			h.BroadcastEvent(emitted, 0)
		// Err() channel will be closed when unsubscribing.
		case <-h.emittedEventsSub.Err():
			return
		}
	}
}

func (h *handler) broadcastProgress() {
	progress := h.myProgress()
	for _, peer := range h.peers.List() {
		peer.AsyncSendProgress(progress, peer.queue)
	}
}

// Progress broadcast loop
func (h *handler) progressBroadcastLoop() {
	ticker := time.NewTicker(h.config.Protocol.ProgressBroadcastPeriod)
	defer ticker.Stop()
	defer h.loopsWg.Done()
	// automatically stops if unsubscribe
	for {
		select {
		case <-ticker.C:
			h.broadcastProgress()
		case <-h.quitProgressBradcast:
			return
		}
	}
}

func (h *handler) onNewEpochLoop() {
	defer h.loopsWg.Done()
	for {
		select {
		case myEpoch := <-h.newEpochsCh:
			h.dagProcessor.Clear()
			h.dagLeecher.OnNewEpoch(myEpoch)
		// Err() channel will be closed when unsubscribing.
		case <-h.newEpochsSub.Err():
			return
		}
	}
}

func (h *handler) txBroadcastLoop() {
	ticker := time.NewTicker(h.config.Protocol.RandomTxHashesSendPeriod)
	defer ticker.Stop()
	defer h.loopsWg.Done()
	for {
		select {
		case notify := <-h.txsCh:
			h.BroadcastTxs(notify.Txs)

		// Err() channel will be closed when unsubscribing.
		case <-h.txsSub.Err():
			return

		case <-ticker.C:
			if !h.syncStatus.AcceptTxs() {
				break
			}
			peers := h.peers.List()
			if len(peers) == 0 {
				continue
			}
			randPeer := peers[rand.Intn(len(peers))]
			h.syncTransactions(randPeer, h.txpool.SampleHashes(h.config.Protocol.MaxRandomTxHashesSend))
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
func (h *handler) NodeInfo() *NodeInfo {
	numOfBlocks := h.store.GetLatestBlockIndex()
	return &NodeInfo{
		Network:     h.NetworkID,
		Genesis:     common.Hash(*h.store.GetGenesisID()),
		Epoch:       h.store.GetEpoch(),
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
