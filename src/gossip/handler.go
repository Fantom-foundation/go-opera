package gossip

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/src/gossip/fetcher"
	"github.com/Fantom-foundation/go-lachesis/src/gossip/ordering"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
)

const (
	softResponseLimit = 2 * 1024 * 1024 // Target maximum size of returned events, or other data.

	// txChanSize is the size of channel listening to NewTxsEvent.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096

	// minimim number of peers to broadcast new events to
	minBroadcastPeers = 4
)

var (
	syncChallengeTimeout = 15 * time.Second // Time allowance for a node to reply to the sync progress challenge
)

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

type ProtocolManager struct {
	config *lachesis.Net

	networkID uint64

	fastSync  uint32 // Flag whether fast sync is enabled (gets disabled if we already have events)
	acceptTxs uint32 // Flag whether we're considered synchronised (enables transaction processing)

	txpool   txPool
	maxPeers int

	peers *peerSet

	mux    *event.TypeMux
	txsCh  chan core.NewTxsEvent
	txsSub event.Subscription

	fetcher *fetcher.Fetcher

	store    *Store
	engine   Consensus
	engineMu *sync.RWMutex

	emittedEventsSub *event.TypeMuxSubscription

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
	config *lachesis.Net,
	mode downloader.SyncMode,
	networkID uint64,
	mux *event.TypeMux,
	txpool txPool,
	engineMu *sync.RWMutex,
	s *Store,
	engine Consensus,
) (
	*ProtocolManager,
	error,
) {
	// Create the protocol manager with the base fields
	pm := &ProtocolManager{
		config:      config,
		networkID:   networkID,
		mux:         mux,
		txpool:      txpool,
		store:       s,
		engine:      engine,
		peers:       newPeerSet(),
		engineMu:    engineMu,
		newPeerCh:   make(chan *peer),
		noMorePeers: make(chan struct{}),
		txsyncCh:    make(chan *txsync),
		quitSync:    make(chan struct{}),
	}

	pm.fetcher = pm.makeFetcher()

	return pm, nil
}

func (pm *ProtocolManager) makeFetcher() *fetcher.Fetcher {
	// build EventBuffer
	pushInBuffer, isEventBuffered := ordering.EventBuffer(ordering.Callback{
		Process: func(e *inter.Event) error {
			log.Info("New event", "hash", e.Hash())

			err := pm.engine.ProcessEvent(e)
			if err != nil {
				return err
			}

			// If the event is indeed in our own graph, announce it
			pm.BroadcastEvent(e, false) // TODO do not announce if it's "initial events download"
			return nil
		},

		Drop: func(e *inter.Event, peer string, err error) {
			log.Warn("Event rejected", "err", err)
			pm.store.DeleteEvent(e.Epoch, e.Hash())
			pm.removePeer(peer)
		},

		Exists: func(id hash.Event) *inter.Event {
			return pm.store.GetEvent(id)
		},
	})

	pushEvent := func(e *inter.Event, peer string) {
		pm.engineMu.Lock()
		defer pm.engineMu.Unlock()

		pushInBuffer(e, peer)
	}

	isEventDownloaded := func(id hash.Event) bool {
		pm.engineMu.RLock()
		defer pm.engineMu.RUnlock()

		if isEventBuffered(id) {
			return true
		}
		return pm.store.HasEvent(id)
	}

	return fetcher.New(pushEvent, isEventDownloaded, pm.removePeer)
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
			peer := pm.newPeer(int(version), p, rw)
			select {
			case pm.newPeerCh <- peer:
				pm.wg.Add(1)
				defer pm.wg.Done()
				return pm.handle(peer)
			case <-pm.quitSync:
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
	//pm.downloader.UnregisterPeer(id)
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
	pm.txsCh = make(chan core.NewTxsEvent, txChanSize)
	pm.txsSub = pm.txpool.SubscribeNewTxsEvent(pm.txsCh)
	go pm.txBroadcastLoop()

	// broadcast mined events
	pm.emittedEventsSub = pm.mux.Subscribe(&inter.Event{})
	go pm.emittedBroadcastLoop()

	// start sync handlers
	go pm.syncer()
	go pm.txsyncLoop()
}

func (pm *ProtocolManager) Stop() {
	log.Info("Stopping Fantom protocol")

	pm.txsSub.Unsubscribe()           // quits txBroadcastLoop
	pm.emittedEventsSub.Unsubscribe() // quits eventBroadcastLoop

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
		myProgress = PeerProgress{
			Epoch:       pm.engine.CurrentSuperFrameN(),
			NumOfEvents: idx.Event(0), // TODO
		}
	)
	if err := p.Handshake(pm.networkID, myProgress, genesis); err != nil {
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

	// Register the peer in the downloader. If the downloader considers it banned, we disconnect
	/*if err := pm.downloader.RegisterPeer(p.id, p.version, p); err != nil {
		return err
	}*/
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

	// Handle the message depending on its contents
	switch {
	case msg.Code == StatusMsg:
		// Status messages should never arrive after the handshake
		return errResp(ErrExtraStatusMsg, "uncontrolled status message")

	case msg.Code == NewEventHashesMsg:
		var announces []hash.Event
		if err := msg.Decode(&announces); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		// Mark the hashes as present at the remote node
		for _, id := range announces {
			p.MarkEvent(id)
		}
		// Schedule all the unknown hashes for retrieval
		_ = pm.fetcher.Notify(p.id, announces, time.Now(), p.RequestEvents)

	case msg.Code == EventsMsg:
		var events []*inter.Event
		if err := msg.Decode(&events); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		// Mark the hashes as present at the remote node
		for _, event := range events {
			p.MarkEvent(event.Hash())
		}

		_ = pm.fetcher.Enqueue(p.id, events)

	case msg.Code == TxMsg:
		// Transactions arrived, make sure we have a valid and fresh chain to handle them
		if atomic.LoadUint32(&pm.acceptTxs) == 0 {
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
		var requests []hash.Event
		if err := msg.Decode(&requests); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}

		rawEvents := make([]rlp.RawValue, 0, len(requests))
		ids := make([]hash.Event, 0, len(requests))
		size := 0
		for _, id := range requests {
			if e := pm.store.GetEventRLP(id); e != nil {
				rawEvents = append(rawEvents, e)
				ids = append(ids, id)
				size += len(e)
			}
			if size > softResponseLimit {
				break
			}
		}
		if len(rawEvents) != 0 {
			_ = p.SendEventsRLP(rawEvents, ids)
		}

	case msg.Code == GetEventHeadersMsg:
		return errResp(ErrExtraStatusMsg, "not supported yet")

	case msg.Code == EventHeadersMsg:
		return errResp(ErrExtraStatusMsg, "not supported yet")

	case msg.Code == GetEventBodiesMsg:
		return errResp(ErrExtraStatusMsg, "not supported yet")

	case msg.Code == EventBodiesMsg:
		return errResp(ErrExtraStatusMsg, "not supported yet")

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
			peer.AsyncSendNewEvent(event)
		}
		log.Trace("Propagated event", "hash", id, "recipients", len(transfer))
		return transferLen
	}
	// Announce it
	for _, peer := range peers {
		peer.AsyncSendNewEventHash(event)
	}
	log.Trace("Announced event", "hash", id, "recipients", len(peers))
	return len(peers)
}

// BroadcastTxs will propagate a batch of transactions to all peers which are not known to
// already have the given transaction.
func (pm *ProtocolManager) BroadcastTxs(txs types.Transactions) {
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
	// automatically stops if unsubscribe
	for obj := range pm.emittedEventsSub.Chan() {
		if ev, ok := obj.Data.(*inter.Event); ok {
			if pm.config.Gossip.ForcedBroadcast {
				pm.BroadcastEvent(ev, true) // No one knows the event, so be aggressive
			}
			pm.BroadcastEvent(ev, false) // Only then announce to the rest
		}
	}
}

func (pm *ProtocolManager) txBroadcastLoop() {
	for {
		select {
		case event := <-pm.txsCh:
			pm.BroadcastTxs(event.Txs)

		// Err() channel will be closed when unsubscribing.
		case <-pm.txsSub.Err():
			return
		}
	}
}

// NodeInfo represents a short summary of the sub-protocol metadata
// known about the host peer.
type NodeInfo struct {
	Network     uint64    `json:"network"` // network ID
	Genesis     hash.Hash `json:"genesis"` // SHA3 hash of the host's genesis object
	Epoch       idx.SuperFrame
	NumOfEvents idx.Event
	//Config  *params.ChainConfig `json:"config"`  // Chain configuration for the fork rules
}

// NodeInfo retrieves some protocol metadata about the running host node.
func (pm *ProtocolManager) NodeInfo() *NodeInfo {
	return &NodeInfo{
		Network: pm.networkID,
		Genesis: pm.engine.GetGenesisHash(),
		Epoch:   pm.engine.CurrentSuperFrameN(),
	}
}
