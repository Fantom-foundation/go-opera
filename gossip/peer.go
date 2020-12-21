package gossip

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/Fantom-foundation/lachesis-base/gossip/dagstream"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/utils/datasemaphore"
	mapset "github.com/deckarep/golang-set"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/inter"
)

var (
	errClosed            = errors.New("peer set is closed")
	errAlreadyRegistered = errors.New("peer is already registered")
	errNotRegistered     = errors.New("peer is not registered")
)

const (
	maxKnownTxs    = 24576 // Maximum transactions hashes to keep in the known list (prevent DOS)
	maxKnownEvents = 24576 // Maximum event hashes to keep in the known list (prevent DOS)

	// maxQueuedItems is the maximum number of items to queue up before
	// dropping broadcasts. This is a sensitive number as a transaction list might
	// contain a single transaction, or thousands.
	maxOrderedQueueItems   = 4096
	maxUnorderedQueueItems = 4096
	maxQueuedSize          = protocolMaxMsgSize + 1024

	handshakeTimeout = 5 * time.Second
)

// PeerInfo represents a short summary of the sub-protocol metadata known
// about a connected peer.
type PeerInfo struct {
	Version     int       `json:"version"` // protocol version negotiated
	Epoch       idx.Epoch `json:"epoch"`
	NumOfBlocks idx.Block `json:"blocks"`
}

type broadcastItem struct {
	Code uint64
	Raw  rlp.RawValue
}

type peer struct {
	id string

	*p2p.Peer
	rw p2p.MsgReadWriter

	version int // Protocol version negotiated

	knownTxs            mapset.Set         // Set of transaction hashes known to be known by this peer
	knownEvents         mapset.Set         // Set of event hashes known to be known by this peer
	unorderedQueue      chan broadcastItem // queue of items to send
	orderedQueue        chan broadcastItem // queue of items to send
	queuedDataSemaphore *datasemaphore.DataSemaphore
	term                chan struct{} // Termination channel to stop the broadcaster

	progress PeerProgress

	poolEntry *poolEntry

	sync.RWMutex
}

func (p *peer) SetProgress(x PeerProgress) {
	p.Lock()
	defer p.Unlock()

	p.progress = x
}

func (p *peer) InterestedIn(h hash.Event) bool {
	e := h.Epoch()

	p.RLock()
	defer p.RUnlock()

	return e != 0 &&
		p.progress.Epoch != 0 &&
		(e == p.progress.Epoch || e == p.progress.Epoch+1) &&
		!p.knownEvents.Contains(h)
}

func (a *PeerProgress) Less(b PeerProgress) bool {
	if a.Epoch != b.Epoch {
		return a.Epoch < b.Epoch
	}
	return a.NumOfBlocks < b.NumOfBlocks
}

func newPeer(version int, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	warningFn := func(received dag.Metric, processing dag.Metric, releasing dag.Metric) {
		log.Warn("Peer queue semaphore inconsistency",
			"receivedNum", received.Num, "receivedSize", received.Size,
			"processingNum", processing.Num, "processingSize", processing.Size,
			"releasingNum", releasing.Num, "releasingSize", releasing.Size)
	}
	return &peer{
		Peer:                p,
		rw:                  rw,
		version:             version,
		id:                  fmt.Sprintf("%x", p.ID().Bytes()[:8]),
		knownTxs:            mapset.NewSet(),
		knownEvents:         mapset.NewSet(),
		unorderedQueue:      make(chan broadcastItem, maxUnorderedQueueItems),
		orderedQueue:        make(chan broadcastItem, maxOrderedQueueItems),
		queuedDataSemaphore: datasemaphore.New(dag.Metric{maxUnorderedQueueItems + maxOrderedQueueItems, maxQueuedSize}, warningFn),
		term:                make(chan struct{}),
	}
}

// broadcast is a write loop that multiplexes event propagations, announcements
// and transaction broadcasts into the remote peer. The goal is to have an async
// writer that does not lock up node internals.
func (p *peer) broadcast(queue chan broadcastItem) {
	for {
		select {
		case item := <-queue:
			_ = p2p.Send(p.rw, item.Code, item.Raw)
			p.queuedDataSemaphore.Release(memSize(item.Raw))

		case <-p.term:
			return
		}
	}
}

// close signals the broadcast goroutine to terminate.
func (p *peer) close() {
	p.queuedDataSemaphore.Terminate()
	close(p.term)
}

// Info gathers and returns a collection of metadata known about a peer.
func (p *peer) Info() *PeerInfo {
	return &PeerInfo{
		Version:     p.version,
		Epoch:       p.progress.Epoch,
		NumOfBlocks: p.progress.NumOfBlocks,
	}
}

// MarkEvent marks a event as known for the peer, ensuring that the event will
// never be propagated to this particular peer.
func (p *peer) MarkEvent(hash hash.Event) {
	// If we reached the memory allowance, drop a previously known event hash
	for p.knownEvents.Cardinality() >= maxKnownEvents {
		p.knownEvents.Pop()
	}
	p.knownEvents.Add(hash)
}

// MarkTransaction marks a transaction as known for the peer, ensuring that it
// will never be propagated to this particular peer.
func (p *peer) MarkTransaction(hash common.Hash) {
	// If we reached the memory allowance, drop a previously known transaction hash
	for p.knownTxs.Cardinality() >= maxKnownTxs {
		p.knownTxs.Pop()
	}
	p.knownTxs.Add(hash)
}

// SendTransactions sends transactions to the peer and includes the hashes
// in its transaction hash set for future reference.
func (p *peer) SendTransactions(txs types.Transactions) error {
	// Mark all the transactions as known, but ensure we don't overflow our limits
	for _, tx := range txs {
		p.knownTxs.Add(tx.Hash())
	}
	for p.knownTxs.Cardinality() >= maxKnownTxs {
		p.knownTxs.Pop()
	}
	return p2p.Send(p.rw, EvmTxMsg, txs)
}

func memSize(v rlp.RawValue) dag.Metric {
	return dag.Metric{1, uint64(len(v) + 1024)}
}

func (p *peer) asyncSendEncodedItem(raw rlp.RawValue, code uint64, queue chan broadcastItem) bool {
	if !p.queuedDataSemaphore.TryAcquire(memSize(raw)) {
		return false
	}
	item := broadcastItem{
		Code: code,
		Raw:  raw,
	}
	select {
	case queue <- item:
		return true
	case <-p.term:
	default:
	}
	p.queuedDataSemaphore.Release(memSize(raw))
	return false
}

func (p *peer) asyncSendNonEncodedItem(value interface{}, code uint64, queue chan broadcastItem) bool {
	raw, err := rlp.EncodeToBytes(value)
	if err != nil {
		return false
	}
	return p.asyncSendEncodedItem(raw, code, queue)
}

func (p *peer) enqueueSendEncodedItem(raw rlp.RawValue, code uint64, queue chan broadcastItem) {
	if !p.queuedDataSemaphore.Acquire(memSize(raw), 10*time.Second) {
		return
	}
	item := broadcastItem{
		Code: code,
		Raw:  raw,
	}
	select {
	case queue <- item:
		return
	case <-p.term:
	}
	p.queuedDataSemaphore.Release(memSize(raw))
}

func (p *peer) enqueueSendNonEncodedItem(value interface{}, code uint64, queue chan broadcastItem) {
	raw, err := rlp.EncodeToBytes(value)
	if err != nil {
		return
	}
	p.enqueueSendEncodedItem(raw, code, queue)
}

func (p *peer) asyncSendTransactions(txs types.Transactions, queue chan broadcastItem) {
	if p.asyncSendNonEncodedItem(txs, EvmTxMsg, queue) {
		// Mark all the transactions as known, but ensure we don't overflow our limits
		for _, tx := range txs {
			p.knownTxs.Add(tx.Hash())
		}
		for p.knownTxs.Cardinality() >= maxKnownTxs {
			p.knownTxs.Pop()
		}
	} else {
		p.Log().Debug("Dropping transactions propagation", "count", len(txs))
	}
}

// AsyncSendTransactions queues list of transactions propagation to a remote
// peer. If the peer's broadcast queue is full, the transactions are silently dropped.
func (p *peer) AsyncSendTransactions(txs types.Transactions, queue chan broadcastItem) {
	// divide big batch into smaller ones
	for len(txs) > 0 {
		batchSize := 0
		var batch types.Transactions
		for i, tx := range txs {
			batchSize += int(tx.Size()) + 1024
			batch = txs[:i+1]
			if batchSize >= softResponseLimitSize || i+1 >= softLimitItems {
				break
			}
		}
		txs = txs[len(batch):]
		p.asyncSendTransactions(batch, queue)
	}
}

// SendEventIDs announces the availability of a number of events through
// a hash notification.
func (p *peer) SendEventIDs(hashes []hash.Event) error {
	// Mark all the event hashes as known, but ensure we don't overflow our limits
	for _, hash := range hashes {
		p.knownEvents.Add(hash)
	}
	for p.knownEvents.Cardinality() >= maxKnownEvents {
		p.knownEvents.Pop()
	}
	return p2p.Send(p.rw, NewEventIDsMsg, hashes)
}

// AsyncSendNewEventHash queues the availability of a event for propagation to a
// remote peer. If the peer's broadcast queue is full, the event is silently
// dropped.
func (p *peer) AsyncSendEventIDs(ids hash.Events, queue chan broadcastItem) {
	if p.asyncSendNonEncodedItem(ids, NewEventIDsMsg, queue) {
		// Mark all the event hash as known, but ensure we don't overflow our limits
		for _, id := range ids {
			p.knownEvents.Add(id)
		}
		for p.knownEvents.Cardinality() >= maxKnownEvents {
			p.knownEvents.Pop()
		}
	} else {
		p.Log().Debug("Dropping event announcement", "count", len(ids))
	}
}

// SendEvents propagates a batch of events to a remote peer.
func (p *peer) SendEvents(events inter.EventPayloads) error {
	// Mark all the event hash as known, but ensure we don't overflow our limits
	for _, event := range events {
		p.knownEvents.Add(event.ID())
		for p.knownEvents.Cardinality() >= maxKnownEvents {
			p.knownEvents.Pop()
		}
	}
	return p2p.Send(p.rw, EventsMsg, events)
}

// SendEventsRLP propagates a batch of RLP events to a remote peer.
func (p *peer) SendEventsRLP(events []rlp.RawValue, ids []hash.Event) error {
	// Mark all the event hash as known, but ensure we don't overflow our limits
	for _, id := range ids {
		p.knownEvents.Add(id)
		for p.knownEvents.Cardinality() >= maxKnownEvents {
			p.knownEvents.Pop()
		}
	}
	return p2p.Send(p.rw, EventsMsg, events)
}

// AsyncSendEvents queues an entire event for propagation to a remote peer.
// If the peer's broadcast queue is full, the events are silently dropped.
func (p *peer) AsyncSendEvents(events inter.EventPayloads, queue chan broadcastItem) bool {
	if p.asyncSendNonEncodedItem(events, EventsMsg, queue) {
		// Mark all the event hash as known, but ensure we don't overflow our limits
		for _, event := range events {
			p.knownEvents.Add(event.ID())
		}
		for p.knownEvents.Cardinality() >= maxKnownEvents {
			p.knownEvents.Pop()
		}
		return true
	} else {
		p.Log().Debug("Dropping event propagation", "count", len(events))
		return false
	}
}

// EnqueueSendEventsRLP queues an entire RLP event for propagation to a remote peer.
// The method is blocking in a case if the peer's broadcast queue is full.
func (p *peer) EnqueueSendEventsRLP(events []rlp.RawValue, ids []hash.Event, queue chan broadcastItem) {
	p.enqueueSendNonEncodedItem(events, EventsMsg, queue)
	// Mark all the event hash as known, but ensure we don't overflow our limits
	for _, id := range ids {
		p.knownEvents.Add(id)
	}
	for p.knownEvents.Cardinality() >= maxKnownEvents {
		p.knownEvents.Pop()
	}
}

// AsyncSendProgress queues a progress propagation to a remote peer.
// If the peer's broadcast queue is full, the progress is silently dropped.
func (p *peer) AsyncSendProgress(progress PeerProgress, queue chan broadcastItem) {
	if !p.asyncSendNonEncodedItem(progress, ProgressMsg, queue) {
		p.Log().Debug("Dropping peer progress propagation")
	}
}

func (p *peer) RequestEvents(ids hash.Events) error {
	// divide big batch into smaller ones
	for start := 0; start < len(ids); start += softLimitItems {
		end := len(ids)
		if end > start+softLimitItems {
			end = start + softLimitItems
		}
		p.Log().Debug("Fetching batch of events", "count", len(ids[start:end]))
		err := p2p.Send(p.rw, GetEventsMsg, ids[start:end])
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *peer) SendEventsStream(r dagstream.Response, ids hash.Events) error {
	// Mark all the event hash as known, but ensure we don't overflow our limits
	for _, id := range ids {
		p.knownEvents.Add(id)
		for p.knownEvents.Cardinality() >= maxKnownEvents {
			p.knownEvents.Pop()
		}
	}
	return p2p.Send(p.rw, EventsStreamResponse, r)
}

func (p *peer) RequestEventsStream(r dagstream.Request) error {
	return p2p.Send(p.rw, RequestEventsStream, r)
}

// Handshake executes the protocol handshake, negotiating version number,
// network IDs, difficulties, head and genesis object.
func (p *peer) Handshake(network uint64, progress PeerProgress, genesis common.Hash) error {
	// Send out own handshake in a new thread
	errc := make(chan error, 2)
	var status ethStatusData // safe to read after two values have been received from errc

	go func() {
		// send both EthStatusMsg and ProgressMsg, eth62 clients will understand only status
		err := p2p.Send(p.rw, EthStatusMsg, &ethStatusData{
			ProtocolVersion:   uint32(p.version),
			NetworkID:         network,
			Genesis:           genesis,
			DummyTD:           big.NewInt(int64(progress.NumOfBlocks)), // for ETH clients
			DummyCurrentBlock: common.Hash(progress.LastBlock),
		})
		if err != nil {
			errc <- err
		}
		errc <- p.SendProgress(progress)
	}()
	go func() {
		errc <- p.readStatus(network, &status, genesis)
		// do not expect ProgressMsg here, because eth62 clients won't send it
	}()
	timeout := time.NewTimer(handshakeTimeout)
	defer timeout.Stop()
	for i := 0; i < 2; i++ {
		select {
		case err := <-errc:
			if err != nil {
				return err
			}
		case <-timeout.C:
			return p2p.DiscReadTimeout
		}
	}
	return nil
}

func (p *peer) SendProgress(progress PeerProgress) error {
	return p2p.Send(p.rw, ProgressMsg, progress)
}

func (p *peer) readStatus(network uint64, status *ethStatusData, genesis common.Hash) (err error) {
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Code != EthStatusMsg {
		return errResp(ErrNoStatusMsg, "first msg has code %x (!= %x)", msg.Code, EthStatusMsg)
	}
	if msg.Size > protocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, protocolMaxMsgSize)
	}
	// Decode the handshake and make sure everything matches
	if err := msg.Decode(&status); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	if status.Genesis != genesis {
		return errResp(ErrGenesisMismatch, "%x (!= %x)", status.Genesis[:8], genesis[:8])
	}
	if status.NetworkID != network {
		return errResp(ErrNetworkIDMismatch, "%d (!= %d)", status.NetworkID, network)
	}
	if int(status.ProtocolVersion) != p.version {
		return errResp(ErrProtocolVersionMismatch, "%d (!= %d)", status.ProtocolVersion, p.version)
	}
	return nil
}

// String implements fmt.Stringer.
func (p *peer) String() string {
	return fmt.Sprintf("Peer %s [%s]", p.id,
		fmt.Sprintf("opera/%2d", p.version),
	)
}

// peerSet represents the collection of active peers currently participating in
// the sub-protocol.
type peerSet struct {
	peers  map[string]*peer
	lock   sync.RWMutex
	closed bool
}

// newPeerSet creates a new peer set to track the active participants.
func newPeerSet() *peerSet {
	return &peerSet{
		peers: make(map[string]*peer),
	}
}

// Register injects a new peer into the working set, or returns an error if the
// peer is already known. If a new peer it registered, its broadcast loop is also
// started.
func (ps *peerSet) Register(p *peer) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if ps.closed {
		return errClosed
	}
	if _, ok := ps.peers[p.id]; ok {
		return errAlreadyRegistered
	}
	ps.peers[p.id] = p
	go p.broadcast(p.orderedQueue)
	for i := 0; i < 2; i++ {
		go p.broadcast(p.unorderedQueue)
	}

	return nil
}

// Unregister removes a remote peer from the active set, disabling any further
// actions to/from that particular entity.
func (ps *peerSet) Unregister(id string) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	p, ok := ps.peers[id]
	if !ok {
		return errNotRegistered
	}
	delete(ps.peers, id)
	p.close()

	return nil
}

// Peer retrieves the registered peer with the given id.
func (ps *peerSet) Peer(id string) *peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return ps.peers[id]
}

// Len returns if the current number of peers in the set.
func (ps *peerSet) Len() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return len(ps.peers)
}

// PeersWithoutEvent retrieves a list of peers that do not have a given event in
// their set of known hashes.
func (ps *peerSet) PeersWithoutEvent(e hash.Event) []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if p.InterestedIn(e) {
			list = append(list, p)
		}
	}
	return list
}

func (ps *peerSet) List() []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		list = append(list, p)
	}
	return list
}

// PeersWithoutTx retrieves a list of peers that do not have a given transaction
// in their set of known hashes.
func (ps *peerSet) PeersWithoutTx(hash common.Hash) []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if !p.knownTxs.Contains(hash) {
			list = append(list, p)
		}
	}
	return list
}

// BestPeer retrieves the known peer with the currently highest total difficulty.
func (ps *peerSet) BestPeer() *peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	var (
		bestPeer     *peer
		bestProgress PeerProgress
	)
	for _, p := range ps.peers {
		if bestProgress.Less(p.progress) {
			bestPeer, bestProgress = p, p.progress
		}
	}
	return bestPeer
}

// Close disconnects all peers.
// No new peers can be registered after Close has returned.
func (ps *peerSet) Close() {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for _, p := range ps.peers {
		p.Disconnect(p2p.DiscQuitting)
	}
	ps.closed = true
}
