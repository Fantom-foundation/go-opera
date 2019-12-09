package gossip

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/eventcheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/basiccheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/heavycheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/parentscheck"
	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/gossip/fetcher"
	"github.com/Fantom-foundation/go-lachesis/gossip/ordering"
	"github.com/Fantom-foundation/go-lachesis/gossip/packsdownloader"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

const (
	softResponseLimitSize = 2 * 1024 * 1024    // Target maximum size of returned events, or other data.
	softLimitItems        = 250                // Target maximum number of events or transactions to request/response
	hardLimitItems        = softLimitItems * 4 // Maximum number of events or transactions to request/response

	// txChanSize is the size of channel listening to NewTxsNotify.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096

	// the maximum number of events in the ordering buffer
	eventsBuffSize = 2048

	// minimum number of peers to broadcast new events to
	minBroadcastPeers = 4
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
	SubscribeNewPack(ch chan<- idx.Pack) notify.Subscription
	SubscribeNewEmitted(ch chan<- *inter.Event) notify.Subscription
}

type ProtocolManager struct {
	config *Config

	//fastSync uint32 // Flag whether fast sync is enabled (gets disabled if we already have events)
	synced uint32 // Flag whether we're considered synchronised (enables transaction processing, events broadcasting)

	txpool   txPool
	maxPeers int

	peers *peerSet

	serverPool *serverPool

	txsCh  chan evmcore.NewTxsNotify
	txsSub notify.Subscription

	downloader *packsdownloader.PacksDownloader
	fetcher    *fetcher.Fetcher
	buffer     *ordering.EventBuffer

	store    *Store
	engine   Consensus
	engineMu *sync.RWMutex

	notifier         dagNotifier
	emittedEventsCh  chan *inter.Event
	emittedEventsSub notify.Subscription
	newPacksCh       chan idx.Pack
	newPacksSub      notify.Subscription
	newEpochsCh      chan idx.Epoch
	newEpochsSub     notify.Subscription

	// channels for fetcher, syncer, txsyncLoop
	newPeerCh   chan *peer
	txsyncCh    chan *txsync
	quitSync    chan struct{}
	noMorePeers chan struct{}

	// wait group is used for graceful shutdowns during downloading
	// and processing
	wg sync.WaitGroup
}

// NewProtocolManager returns a new Fantom sub protocol manager. The Fantom sub protocol manages peers capable
// with the Fantom network.
func NewProtocolManager(
	config *Config,
	notifier dagNotifier,
	txpool txPool,
	engineMu *sync.RWMutex,
	s *Store,
	engine Consensus,
	serverPool *serverPool,
) (
	*ProtocolManager,
	error,
) {
	// Create the protocol manager with the base fields
	pm := &ProtocolManager{
		config:      config,
		notifier:    notifier,
		txpool:      txpool,
		store:       s,
		engine:      engine,
		peers:       newPeerSet(),
		serverPool:  serverPool,
		engineMu:    engineMu,
		newPeerCh:   make(chan *peer),
		noMorePeers: make(chan struct{}),
		txsyncCh:    make(chan *txsync),
		quitSync:    make(chan struct{}),
	}

	pm.fetcher, pm.buffer = pm.makeFetcher()
	pm.downloader = packsdownloader.New(pm.fetcher, pm.onlyNotConnectedEvents, pm.removePeer)

	return pm, nil
}

func (pm *ProtocolManager) makeFetcher() (*fetcher.Fetcher, *ordering.EventBuffer) {
	// checkers
	basicCheck := basiccheck.New(&pm.config.Net.Dag)
	epochCheck := epochcheck.New(&pm.config.Net.Dag, pm.engine)
	parentsCheck := parentscheck.New(&pm.config.Net.Dag)
	firstCheck := func(e *inter.Event) error {
		if err := basicCheck.Validate(e); err != nil {
			return err
		}
		if err := epochCheck.Validate(e); err != nil {
			return err
		}
		return nil
	}

	dagID := params.AllEthashProtocolChanges.ChainID
	txSigner := types.NewEIP155Signer(dagID)
	heavyCheck := heavycheck.NewDefault(&pm.config.Net.Dag, txSigner)

	// DAG callbacks
	buffer := ordering.New(eventsBuffSize, ordering.Callback{

		Process: func(e *inter.Event) error {
			pm.engineMu.Lock()
			defer pm.engineMu.Unlock()

			log.Info("New event", "event", e.String())
			err := pm.engine.ProcessEvent(e)
			if err != nil {
				return err
			}

			// If the event is indeed in our own graph, announce it
			if atomic.LoadUint32(&pm.synced) != 0 { // announce only fresh events
				pm.BroadcastEvent(e, false)
			}

			return nil
		},

		Drop: func(e *inter.Event, peer string, err error) {
			if eventcheck.IsBan(err) {
				log.Warn("Incoming event rejected", "event", e.Hash().String(), "creator", e.Creator.String(), "err", err)
				pm.removePeer(peer)
			}
		},

		Exists: func(id hash.Event) *inter.EventHeaderData {
			return pm.store.GetEventHeader(id.Epoch(), id)
		},

		Check: parentsCheck.Validate,
	})

	newFetcher := fetcher.New(fetcher.Callback{
		PushEvent:      buffer.PushEvent,
		OnlyInterested: pm.onlyInterestedEvents,
		DropPeer:       pm.removePeer,
		FirstCheck:     firstCheck,
		HeavyCheck:     heavyCheck,
	})
	return newFetcher, buffer
}

func (pm *ProtocolManager) onlyNotConnectedEvents(ids hash.Events) hash.Events {
	if len(ids) == 0 {
		return ids
	}

	notConnected := make(hash.Events, 0, len(ids))
	for _, id := range ids {
		if pm.store.HasEvent(id) {
			continue
		}
		notConnected.Add(id)
	}
	return notConnected
}

func (pm *ProtocolManager) onlyInterestedEvents(ids hash.Events) hash.Events {
	if len(ids) == 0 {
		return ids
	}
	epoch := pm.engine.GetEpoch()

	interested := make(hash.Events, 0, len(ids))
	for _, id := range ids {
		if id.Epoch() != epoch {
			continue
		}
		if pm.buffer.IsBuffered(id) || pm.store.HasEvent(id) {
			continue
		}
		interested.Add(id)
	}
	return interested
}

func (pm *ProtocolManager) makeProtocol(version uint) p2p.Protocol {
	length, ok := protocolLengths[version]
	if !ok {
		panic("makeProtocol for unknown version")
	}

	return p2p.Protocol{
		Name:    protocolName,
		Version: version,
		Length:  length,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			var entry *poolEntry
			peer := pm.newPeer(int(version), p, rw)
			if pm.serverPool != nil {
				entry = pm.serverPool.connect(peer, peer.Node())
			}
			peer.poolEntry = entry
			select {
			case pm.newPeerCh <- peer:
				pm.wg.Add(1)
				defer pm.wg.Done()
				err := pm.handle(peer)
				if entry != nil {
					pm.serverPool.disconnect(entry)
				}
				return err
			case <-pm.quitSync:
				if entry != nil {
					pm.serverPool.disconnect(entry)
				}
				return p2p.DiscQuitting
			}
		},
		NodeInfo: func() interface{} {
			return pm.NodeInfo()
		},
		PeerInfo: func(id enode.ID) interface{} {
			if p := pm.peers.Peer(fmt.Sprintf("%x", id[:8])); p != nil {
				return p.Info()
			}
			return nil
		},
	}
}

func (pm *ProtocolManager) removePeer(id string) {
	// Short circuit if the peer was already removed
	peer := pm.peers.Peer(id)
	if peer == nil {
		return
	}
	log.Debug("Removing peer", "peer", id)

	// Unregister the peer from the downloader and peer set
	_ = pm.downloader.UnregisterPeer(id)
	if err := pm.peers.Unregister(id); err != nil {
		log.Error("Peer removal failed", "peer", id, "err", err)
	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
	}
}

func (pm *ProtocolManager) Start(maxPeers int) {
	pm.maxPeers = maxPeers

	// broadcast transactions
	pm.txsCh = make(chan evmcore.NewTxsNotify, txChanSize)
	pm.txsSub = pm.txpool.SubscribeNewTxsNotify(pm.txsCh)
	go pm.txBroadcastLoop()

	if pm.notifier != nil {
		// broadcast mined events
		pm.emittedEventsCh = make(chan *inter.Event, 4)
		pm.emittedEventsSub = pm.notifier.SubscribeNewEmitted(pm.emittedEventsCh)
		// broadcast packs
		pm.newPacksCh = make(chan idx.Pack, 4)
		pm.newPacksSub = pm.notifier.SubscribeNewPack(pm.newPacksCh)
		// epoch changes
		pm.newEpochsCh = make(chan idx.Epoch, 4)
		pm.newEpochsSub = pm.notifier.SubscribeNewEpoch(pm.newEpochsCh)
	}

	go pm.emittedBroadcastLoop()
	go pm.progressBroadcastLoop()
	go pm.onNewEpochLoop()

	// start sync handlers
	go pm.syncer()
	go pm.txsyncLoop()
}

func (pm *ProtocolManager) Stop() {
	log.Info("Stopping Fantom protocol")

	pm.txsSub.Unsubscribe() // quits txBroadcastLoop
	if pm.notifier != nil {
		pm.emittedEventsSub.Unsubscribe() // quits eventBroadcastLoop
		pm.newPacksSub.Unsubscribe()      // quits progressBroadcastLoop
		pm.newEpochsSub.Unsubscribe()     // quits onNewEpochLoop
	}

	// Quit the sync loop.
	// After this send has completed, no new peers will be accepted.
	pm.noMorePeers <- struct{}{}

	// Quit fetcher, txsyncLoop.
	close(pm.quitSync)

	// Disconnect existing sessions.
	// This also closes the gate for any new registrations on the peer set.
	// sessions which are already established but not added to pm.peers yet
	// will exit when they try to register.
	pm.peers.Close()

	// Wait for all peer handler goroutines and the loops to come down.
	pm.wg.Wait()

	log.Info("Fantom protocol stopped")
}

func (pm *ProtocolManager) newPeer(pv int, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return newPeer(pv, p, rw)
}

func (pm *ProtocolManager) myProgress() PeerProgress {
	blockI, block := pm.engine.LastBlock()
	epoch := pm.engine.GetEpoch()
	return PeerProgress{
		Epoch:        epoch,
		NumOfBlocks:  blockI,
		LastBlock:    block,
		LastPackInfo: pm.store.GetPackInfoOrDefault(epoch, pm.store.GetPacksNumOrDefault(epoch)-1),
	}
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
		genesis    = pm.engine.GetGenesisHash()
		myProgress = pm.myProgress()
	)
	if err := p.Handshake(pm.config.Net.NetworkID, myProgress, genesis); err != nil {
		p.Log().Debug("Handshake failed", "err", err)
		return err
	}
	//if rw, ok := p.rw.(*meteredMsgReadWriter); ok {
	//	rw.Init(p.version)
	//}
	// Register the peer locally
	if err := pm.peers.Register(p); err != nil {
		p.Log().Error("Peer registration failed", "err", err)
		return err
	}
	defer pm.removePeer(p.id)

	// Propagate existing transactions. new transactions appearing
	// after this will be sent via broadcasts.
	pm.syncTransactions(p)

	// Handle incoming messages until the connection is torn down
	for {
		if err := pm.handleMsg(p); err != nil {
			p.Log().Debug("Message handling failed", "err", err)
			return err
		}
	}
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

	myEpoch := pm.engine.GetEpoch()
	peerDwnlr := pm.downloader.Peer(p.id)

	// Handle the message depending on its contents
	switch {
	case msg.Code == EthStatusMsg:
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

		// notify downloader about new peer's epoch
		_ = pm.downloader.RegisterPeer(packsdownloader.Peer{
			ID:               p.id,
			Epoch:            p.progress.Epoch,
			RequestPack:      p.RequestPack,
			RequestPackInfos: p.RequestPackInfos,
		}, myEpoch)
		peerDwnlr = pm.downloader.Peer(p.id)

		if peerDwnlr != nil && progress.LastPackInfo.Index > 0 {
			_ = peerDwnlr.NotifyPackInfo(p.progress.Epoch, progress.LastPackInfo.Index, progress.LastPackInfo.Heads, time.Now())
		}

	case msg.Code == NewEventHashesMsg:
		if pm.fetcher.Overloaded() {
			break
		}
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
		// Mark the hashes as present at the remote node
		for _, id := range announces {
			p.MarkEvent(id)
		}
		// Schedule all the unknown hashes for retrieval
		_ = pm.fetcher.Notify(p.id, announces, time.Now(), p.RequestEvents)

	case msg.Code == EventsMsg:
		if pm.fetcher.Overloaded() {
			break
		}
		var events []*inter.Event
		if err := msg.Decode(&events); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if err := checkLenLimits(len(events), events); err != nil {
			return err
		}
		// Mark the hashes as present at the remote node
		for _, e := range events {
			p.MarkEvent(e.Hash())
		}
		_ = pm.fetcher.Enqueue(p.id, events, time.Now(), p.RequestEvents)

	case msg.Code == EvmTxMsg:
		// Transactions arrived, make sure we have a valid and fresh graph to handle them
		if atomic.LoadUint32(&pm.synced) == 0 {
			break
		}
		// Transactions can be processed, parse all of them and deliver to the pool
		var txs []*types.Transaction
		if err := msg.Decode(&txs); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		for i, tx := range txs {
			// Validate and mark the remote transaction
			if tx == nil {
				return errResp(ErrDecode, "transaction %d is nil", i)
			}
			p.MarkTransaction(tx.Hash())
		}
		pm.txpool.AddRemotes(txs)

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
			if raw := pm.store.GetEventRLP(id); raw != nil {
				rawEvents = append(rawEvents, raw)
				ids = append(ids, id)
				size += len(raw)
			}
			if size >= softResponseLimitSize {
				break
			}
		}
		if len(rawEvents) != 0 {
			_ = p.SendEventsRLP(rawEvents, ids)
		}

	case msg.Code == GetPackInfosMsg:
		var request getPackInfosData
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if err := checkLenLimits(len(request.Indexes), request); err != nil {
			return err
		}

		packsNum, ok := pm.store.GetPacksNum(request.Epoch)
		if !ok {
			// no packs in the requested epoch
			break
		}

		rawPackInfos := make([]rlp.RawValue, 0, len(request.Indexes))
		size := 0
		for _, index := range request.Indexes {
			if index >= packsNum {
				// return only pinned and existing packs
				continue
			}

			if raw := pm.store.GetPackInfoRLP(request.Epoch, index); raw != nil {
				rawPackInfos = append(rawPackInfos, raw)
				size += len(raw)
			}
			if size >= softResponseLimitSize {
				break
			}
		}
		if len(rawPackInfos) != 0 {
			_ = p.SendPackInfosRLP(&packInfosDataRLP{
				Epoch:           request.Epoch,
				TotalNumOfPacks: packsNum,
				RawInfos:        rawPackInfos,
			})
		}

	case msg.Code == GetPackMsg:
		var request getPackData
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}

		if request.Epoch > myEpoch {
			// short circuit if future epoch
			break
		}

		ids := make(hash.Events, 0, softLimitItems)
		for i, id := range pm.store.GetPack(request.Epoch, request.Index) {
			ids = append(ids, id)
			if i >= softLimitItems {
				break
			}
		}
		if len(ids) != 0 {
			_ = p.SendPack(&packData{
				Epoch: request.Epoch,
				Index: request.Index,
				IDs:   ids,
			})
		}

	case msg.Code == PackInfosMsg:
		if peerDwnlr == nil {
			break
		}

		var infos packInfosData
		if err := msg.Decode(&infos); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if err := checkLenLimits(len(infos.Infos), infos); err != nil {
			return err
		}

		// notify about number of packs this peer has
		_ = peerDwnlr.NotifyPacksNum(infos.Epoch, infos.TotalNumOfPacks)

		for _, info := range infos.Infos {
			if len(info.Heads) == 0 {
				return errResp(ErrEmptyMessage, "%v", msg)
			}
			// Mark the hashes as present at the remote node
			for _, id := range info.Heads {
				p.MarkEvent(id)
			}
			// Notify downloader about new packInfo
			_ = peerDwnlr.NotifyPackInfo(infos.Epoch, info.Index, info.Heads, time.Now())
		}

	case msg.Code == PackMsg:
		if peerDwnlr == nil {
			break
		}

		var pack packData
		if err := msg.Decode(&pack); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		if err := checkLenLimits(len(pack.IDs), pack); err != nil {
			return err
		}
		if len(pack.IDs) == 0 {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		// Mark the hashes as present at the remote node
		for _, id := range pack.IDs {
			p.MarkEvent(id)
		}
		// Notify downloader about new pack
		_ = peerDwnlr.NotifyPack(pack.Epoch, pack.Index, pack.IDs, time.Now(), p.RequestEvents)

	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}

// BroadcastEvent will either propagate a event to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
func (pm *ProtocolManager) BroadcastEvent(event *inter.Event, aggressive bool) int {
	id := event.Hash()
	peers := pm.peers.PeersWithoutEvent(id)

	// If propagation is requested, send to a subset of the peer
	if aggressive {
		// Send the event to a subset of our peers
		transferLen := int(math.Sqrt(float64(len(peers))))
		if transferLen < minBroadcastPeers {
			transferLen = minBroadcastPeers
		}
		if transferLen > len(peers) {
			transferLen = len(peers)
		}
		transfer := peers[:transferLen]
		for _, peer := range transfer {
			peer.AsyncSendEvents(inter.Events{event})
		}
		log.Trace("Propagated event", "hash", id, "recipients", len(transfer))
		return transferLen
	}
	// Announce it
	for _, peer := range peers {
		peer.AsyncSendNewEventHashes(hash.Events{event.Hash()})
	}
	log.Trace("Announced event", "hash", id, "recipients", len(peers))
	return len(peers)
}

// BroadcastTxs will propagate a batch of transactions to all peers which are not known to
// already have the given transaction.
func (pm *ProtocolManager) BroadcastTxs(txs types.Transactions) {
	if len(txs) > softLimitItems {
		txs = txs[:softLimitItems]
	}

	var txset = make(map[*peer]types.Transactions)

	// Broadcast transactions to a batch of peers not knowing about it
	for _, tx := range txs {
		peers := pm.peers.PeersWithoutTx(tx.Hash())
		for _, peer := range peers {
			txset[peer] = append(txset[peer], tx)
		}
		log.Trace("Broadcast transaction", "hash", tx.Hash(), "recipients", len(peers))
	}
	// FIXME include this again: peers = peers[:int(math.Sqrt(float64(len(peers))))]
	for peer, txs := range txset {
		peer.AsyncSendTransactions(txs)
	}
}

// Mined broadcast loop
func (pm *ProtocolManager) emittedBroadcastLoop() {
	for {
		select {
		case emitted := <-pm.emittedEventsCh:
			if pm.config.ForcedBroadcast {
				pm.BroadcastEvent(emitted, true) // No one knows the event, so be aggressive
			}
			pm.BroadcastEvent(emitted, false) // Only then announce to the rest
		// Err() channel will be closed when unsubscribing.
		case <-pm.txsSub.Err():
			return
		}
	}
}

// Progress broadcast loop
func (pm *ProtocolManager) progressBroadcastLoop() {
	// automatically stops if unsubscribe
	prevProgress := pm.myProgress()
	for {
		select {
		case _ = <-pm.newPacksCh:
			// broadcast my new progress, but not recent one,
			// so others could receive all the events before this node announces the pack
			for _, peer := range pm.peers.List() {
				err := peer.SendProgress(prevProgress)
				if err != nil {
					log.Error("Failed to send progress status", "peer", peer.id)
				}
			}
			prevProgress = pm.myProgress()
		// Err() channel will be closed when unsubscribing.
		case <-pm.txsSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) onNewEpochLoop() {
	for {
		select {
		case myEpoch := <-pm.newEpochsCh:
			peerEpoch := func(peer string) idx.Epoch {
				p := pm.peers.Peer(peer)
				if p == nil {
					return 0
				}
				return p.progress.Epoch
			}
			for _, peer := range pm.peers.List() {
				if peer.progress.Epoch == myEpoch {
					atomic.StoreUint32(&pm.synced, 1) // Mark initial sync done on any peer which has the same epoch
				}
			}
			pm.buffer.Clear()
			pm.downloader.OnNewEpoch(myEpoch, peerEpoch)
		// Err() channel will be closed when unsubscribing.
		case <-pm.txsSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) txBroadcastLoop() {
	for {
		select {
		case notify := <-pm.txsCh:
			pm.BroadcastTxs(notify.Txs)

		// Err() channel will be closed when unsubscribing.
		case <-pm.txsSub.Err():
			return
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
	numOfBlocks, _ := pm.engine.LastBlock()
	return &NodeInfo{
		Network:     pm.config.Net.NetworkID,
		Genesis:     pm.engine.GetGenesisHash(),
		Epoch:       pm.engine.GetEpoch(),
		NumOfBlocks: numOfBlocks,
	}
}
