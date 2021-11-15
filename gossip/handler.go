package gossip

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Fantom-foundation/lachesis-base/eventcheck/queuedcheck"
	"github.com/Fantom-foundation/lachesis-base/gossip/dagprocessor"
	"github.com/Fantom-foundation/lachesis-base/gossip/dagstream"
	"github.com/Fantom-foundation/lachesis-base/gossip/dagstream/streamleecher"
	"github.com/Fantom-foundation/lachesis-base/gossip/dagstream/streamseeder"
	"github.com/Fantom-foundation/lachesis-base/gossip/itemsfetcher"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/utils/datasemaphore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/event"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/Fantom-foundation/go-opera/eventcheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/parentlesscheck"
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/inter"
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

// handlerConfig is the collection of initialization parameters to create a full
// node network handler.
type handlerConfig struct {
	config       Config
	notifier     dagNotifier
	EventMux     *event.TypeMux
	txpool       txPool
	engineMu     sync.Locker
	checkers     *eventcheck.Checkers
	s            *Store
	processEvent func(*inter.EventPayload) error
}

type handler struct {
	NetworkID uint64
	config    Config

	synced uint32 // Flag whether we're considered synchronised (enables transaction processing, events broadcasting)

	txpool   txPool
	maxPeers int

	peers *peerSet

	txsCh  chan evmcore.NewTxsNotify
	txsSub notify.Subscription

	leecher    *streamleecher.Leecher
	seeder     *streamseeder.Seeder
	dagFetcher *itemsfetcher.Fetcher
	txFetcher  *itemsfetcher.Fetcher
	processor  *dagprocessor.Processor
	checkers   *eventcheck.Checkers

	msgSemaphore *datasemaphore.DataSemaphore

	store        *Store
	processEvent func(*inter.EventPayload) error
	engineMu     sync.Locker

	notifier             dagNotifier
	emittedEventsCh      chan *inter.EventPayload
	emittedEventsSub     notify.Subscription
	newEpochsCh          chan idx.Epoch
	newEpochsSub         notify.Subscription
	quitProgressBradcast chan struct{}

	// channels for syncer, txsyncLoop
	txsyncCh chan *txsync
	quitSync chan struct{}

	// geth fields
	chain            *ethBlockChain
	fastSync         uint32      // Flag whether fast sync is enabled (gets disabled if we already have blocks)
	snapSync         uint32      // Flag whether fast sync should operate on top of the snap protocol
	checkpointNumber uint64      // Block number for the sync progress validator to cross reference
	checkpointHash   common.Hash // Block hash for the sync progress validator to cross reference
	downloader       *downloader.Downloader
	stateBloom       *trie.SyncBloom

	// wait group is used for graceful shutdowns during downloading
	// and processing
	loopsWg sync.WaitGroup
	wg      sync.WaitGroup
	peerWG  sync.WaitGroup

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
		processEvent:         c.processEvent,
		checkers:             c.checkers,
		peers:                newPeerSet(),
		engineMu:             c.engineMu,
		txsyncCh:             make(chan *txsync),
		quitSync:             make(chan struct{}),
		quitProgressBradcast: make(chan struct{}),

		Instance: logger.New("PM"),
	}

	// TODO: configure it
	var (
		configSync                                 = downloader.FullSync
		configCheckpoint *params.TrustedCheckpoint = nil
		configBloomCache uint64                    = 0 // Megabytes to alloc for fast sync bloom
	)

	var err error
	h.chain, err = newEthBlockChain(c.s)
	if err != nil {
		return nil, err
	}

	if configSync == downloader.FullSync {
		// The database seems empty as the current block is the genesis. Yet the fast
		// block is ahead, so fast sync was enabled for this node at a certain point.
		// The scenarios where this can happen is
		// * if the user manually (or via a bad block) rolled back a fast sync node
		//   below the sync point.
		// * the last fast sync is not finished while user specifies a full sync this
		//   time. But we don't have any recent state for full sync.
		// In these cases however it's safe to reenable fast sync.
		fullBlock, fastBlock := h.chain.CurrentBlock(), h.chain.CurrentFastBlock()
		if fullBlock.NumberU64() == 0 && fastBlock.NumberU64() > 0 {
			h.fastSync = uint32(1)
			log.Warn("Switch sync mode from full sync to fast sync")
		}
	} else {
		if h.chain.CurrentBlock().NumberU64() > 0 {
			// Print warning log if database is not empty to run fast sync.
			log.Warn("Switch sync mode from fast sync to full sync")
		} else {
			// If fast sync was requested and our database is empty, grant it
			h.fastSync = uint32(1)
			if configSync == downloader.SnapSync {
				h.snapSync = uint32(1)
			}
		}
	}
	// If we have trusted checkpoints, enforce them on the chain
	if configCheckpoint != nil {
		h.checkpointNumber = (configCheckpoint.SectionIndex+1)*params.CHTFrequency - 1
		h.checkpointHash = configCheckpoint.SectionHead
	}
	if c.EventMux == nil {
		c.EventMux = new(event.TypeMux) // Nicety initialization for tests.
	}
	// Construct the downloader (long sync) and its backing state bloom if fast
	// sync is requested. The downloader is responsible for deallocating the state
	// bloom when it's done.
	// Note: we don't enable it if snap-sync is performed, since it's very heavy
	// and the heal-portion of the snap sync is much lighter than fast. What we particularly
	// want to avoid, is a 90%-finished (but restarted) snap-sync to begin
	// indexing the entire trie
	stateDb := h.store.EvmStore().EvmTable()
	if atomic.LoadUint32(&h.fastSync) == 1 && atomic.LoadUint32(&h.snapSync) == 0 {
		h.stateBloom = trie.NewSyncBloom(configBloomCache, stateDb)
	}
	h.downloader = downloader.New(h.checkpointNumber, stateDb, h.stateBloom, c.EventMux, h.chain, nil, h.removePeer)

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
	h.processor = h.makeProcessor(c.checkers)
	h.leecher = streamleecher.New(h.store.GetEpoch(), h.store.GetHighestLamport() == 0, h.config.Protocol.StreamLeecher, streamleecher.Callbacks{
		OnlyNotConnected: h.onlyNotConnectedEvents,
		RequestChunk: func(peer string, r dagstream.Request) error {
			p := h.peers.Peer(peer)
			if p == nil {
				return errNotRegistered
			}
			return p.RequestEventsStream(r)
		},
		Suspend: func(_ string) bool {
			return h.dagFetcher.Overloaded() || h.processor.Overloaded()
		},
		PeerEpoch: func(peer string) idx.Epoch {
			p := h.peers.Peer(peer)
			if p == nil {
				return 0
			}
			return p.progress.Epoch
		},
	})
	h.seeder = streamseeder.New(h.config.Protocol.StreamSeeder, streamseeder.Callbacks{
		ForEachEvent: func(start []byte, onEvent func(key hash.Event, event interface{}, size uint64) bool) {
			c.s.ForEachEventRLP(start, func(key hash.Event, event rlp.RawValue) bool {
				return onEvent(key, event, uint64(len(event)))
			})
		},
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

func (h *handler) makeProcessor(checkers *eventcheck.Checkers) *dagprocessor.Processor {
	// checkers
	lightCheck := func(e dag.Event) error {
		if h.processor.IsBuffered(e.ID()) {
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

	parentlessChecker := parentlesscheck.New(parentlesscheck.Callback{
		OnlyInterested: h.onlyInterestedEvents,
		HeavyCheck:     checkers.Heavycheck,
		LightCheck:     lightCheck,
	})

	newProcessor := dagprocessor.New(datasemaphore.New(h.config.Protocol.EventsSemaphoreLimit, getSemaphoreWarningFn("DAG events")), h.config.Protocol.Processor, dagprocessor.Callback{
		// DAG callbacks
		Event: dagprocessor.EventCallback{
			Process: func(_e dag.Event) error {
				e := _e.(*inter.EventPayload)
				preStart := time.Now()
				h.engineMu.Lock()
				defer h.engineMu.Unlock()

				err := h.processEvent(e)
				if err != nil {
					return err
				}

				// event is connected, announce it if synced up
				if atomic.LoadUint32(&h.synced) != 0 {
					passedSinceEvent := preStart.Sub(e.CreationTime().Time())
					h.BroadcastEvent(e, passedSinceEvent)
				}

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

			CheckParents: bufferedCheck,
			CheckParentless: func(tasks []queuedcheck.EventTask, checked func([]queuedcheck.EventTask)) {
				_ = parentlessChecker.Enqueue(tasks, checked)
			},
			OnlyInterested: h.onlyInterestedEvents,
		},
		PeerMisbehaviour: h.peerMisbehaviour,
		HighestLamport:   h.store.GetHighestLamport,
	})

	return newProcessor
}

func (h *handler) onlyNotConnectedEvents(ids hash.Events) hash.Events {
	if len(ids) == 0 {
		return ids
	}

	notConnected := make(hash.Events, 0, len(ids))
	for _, id := range ids {
		if h.store.HasEvent(id) {
			continue
		}
		notConnected.Add(id)
	}
	return notConnected
}

func (h *handler) isEventInterested(id hash.Event, epoch idx.Epoch) bool {
	if id.Epoch() != epoch {
		return false
	}
	if h.processor.IsBuffered(id) || h.store.HasEvent(id) {
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

func (h *handler) onlyInterestedEvents(ids hash.Events) hash.Events {
	if len(ids) == 0 {
		return ids
	}
	epoch := h.store.GetEpoch()
	interested := make(hash.Events, 0, len(ids))
	for _, id := range ids {
		if h.isEventInterested(id, epoch) {
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
	_ = h.leecher.UnregisterPeer(id)
	_ = h.seeder.UnregisterPeer(id)
	if err := h.peers.UnregisterPeer(id); err != nil {
		log.Error("Peer removal failed", "peer", id, "err", err)
	}
}

func (h *handler) Start(maxPeers int) {
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
	h.dagFetcher.Start()
	h.txFetcher.Start()
	h.checkers.Heavycheck.Start()
	h.processor.Start()
	h.seeder.Start()
	h.leecher.Start()
}

func (h *handler) Stop() {
	log.Info("Stopping Fantom protocol")

	h.leecher.Stop()
	h.seeder.Stop()
	h.processor.Stop()
	h.checkers.Heavycheck.Stop()
	h.txFetcher.Stop()
	h.dagFetcher.Stop()

	close(h.quitProgressBradcast)
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
	// Ignore maxPeers if this is a trusted peer
	if h.peers.Len() >= h.maxPeers && !p.Peer.Info().Network.Trusted {
		return p2p.DiscTooManyPeers
	}
	p.Log().Debug("Peer connected", "name", p.Name())

	// If the peer has a `snap` extension, wait for it to connect so we can have
	// a uniform initialization/teardown mechanism
	snap, err := h.peers.WaitSnapExtension(p)
	if err != nil {
		p.Log().Error("Snapshot extension barrier failed", "err", err)
		return err
	}

	h.peerWG.Add(1)
	defer h.peerWG.Done()

	// Execute the handshake
	var (
		genesis    = *h.store.GetGenesisHash()
		myProgress = h.myProgress()
	)
	if err := p.Handshake(h.NetworkID, myProgress, common.Hash(genesis)); err != nil {
		p.Log().Debug("Handshake failed", "err", err)
		return err
	}

	// Register the peer locally
	if err := h.peers.RegisterPeer(p, snap); err != nil {
		p.Log().Warn("Peer registration failed", "err", err)
		return err
	}
	if err := h.leecher.RegisterPeer(p.id); err != nil {
		p.Log().Warn("Leecher peer registration failed", "err", err)
		return err
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
	sessionCfg := h.config.Protocol.StreamLeecher.Session
	for _, id := range announces {
		maxLamport := h.store.GetHighestLamport() + idx.Lamport(sessionCfg.DefaultChunkSize.Num+1)*idx.Lamport(sessionCfg.ParallelChunksDownload)
		if id.Lamport() <= maxLamport {
			notTooHigh = append(notTooHigh, id)
		}
	}
	if len(announces) != len(notTooHigh) {
		h.leecher.ForceSyncing()
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
	sessionCfg := h.config.Protocol.StreamLeecher.Session
	for _, e := range events {
		maxLamport := h.store.GetHighestLamport() + idx.Lamport(sessionCfg.DefaultChunkSize.Num+1)*idx.Lamport(sessionCfg.ParallelChunksDownload)
		if e.Lamport() <= maxLamport {
			notTooHigh = append(notTooHigh, e)
		}
	}
	if len(events) != len(notTooHigh) {
		h.leecher.ForceSyncing()
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
		_ = h.dagFetcher.NotifyAnnounces(peer.id, eventIDsToInterfaces(ids), now, requestEvents)
	}
	_ = h.processor.Enqueue(peer.id, notTooHigh, ordered, notifyAnnounces, nil)
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

	myEpoch := h.store.GetEpoch()

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
			atomic.StoreUint32(&h.synced, 1) // Mark initial sync done on any peer which has the same epoch
		}

	case msg.Code == EvmTxsMsg:
		// Transactions arrived, make sure we have a valid and fresh graph to handle them
		if atomic.LoadUint32(&h.synced) == 0 {
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
		if atomic.LoadUint32(&h.synced) == 0 {
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
		if atomic.LoadUint32(&h.synced) == 0 {
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

		pid := p.id
		_, peerErr := h.seeder.NotifyRequestReceived(streamseeder.Peer{
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
		var chunk epochChunk
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

		_ = h.leecher.NotifyChunkReceived(chunk.SessionID, last, chunk.Done)

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
			h.processor.Clear()
			if atomic.LoadUint32(&h.synced) == 0 {
				synced := false
				for _, peer := range h.peers.List() {
					if peer.progress.Epoch == myEpoch {
						synced = true
					}
				}
				// Mark initial sync done on any peer which has the same epoch
				if synced {
					atomic.StoreUint32(&h.synced, 1)
				}
			}
			h.leecher.OnNewEpoch(myEpoch)
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
			if atomic.LoadUint32(&h.synced) == 0 {
				continue
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
		Genesis:     common.Hash(*h.store.GetGenesisHash()),
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
