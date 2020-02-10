package gossip

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

var (
	errClosed            = errors.New("peer set is closed")
	errAlreadyRegistered = errors.New("peer is already registered")
	errNotRegistered     = errors.New("peer is not registered")
)

const (
	maxKnownTxs    = 24576 // Maximum transactions hashes to keep in the known list (prevent DOS)
	maxKnownEvents = 16384 // Maximum event hashes to keep in the known list (prevent DOS)

	// maxQueuedTxs is the maximum number of transaction lists to queue up before
	// dropping broadcasts. This is a sensitive number as a transaction list might
	// contain a single transaction, or thousands.
	maxQueuedTxs = 128

	// maxQueuedProps is the maximum number of event propagations to queue up before
	// dropping broadcasts.
	maxQueuedProps = 128

	// maxQueuedAnns is the maximum number of event announcements to queue up before
	// dropping broadcasts.
	maxQueuedAnns = 128

	handshakeTimeout = 5 * time.Second
)

// PeerInfo represents a short summary of the sub-protocol metadata known
// about a connected peer.
type PeerInfo struct {
	Version     int       `json:"version"` // protocol version negotiated
	Epoch       idx.Epoch `json:"epoch"`
	NumOfBlocks idx.Block `json:"blocks"`
}

type peer struct {
	id string

	*p2p.Peer
	rw p2p.MsgReadWriter

	version int // Protocol version negotiated

	knownTxs    mapset.Set                // Set of transaction hashes known to be known by this peer
	knownEvents mapset.Set                // Set of event hashes known to be known by this peer
	queuedTxs   chan []*types.Transaction // Queue of transactions to broadcast to the peer
	queuedProps chan inter.Events         // Queue of events to broadcast to the peer
	queuedAnns  chan hash.Events          // Queue of events to announce to the peer
	term        chan struct{}             // Termination channel to stop the broadcaster

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
	if a.NumOfBlocks != b.NumOfBlocks {
		return a.NumOfBlocks < b.NumOfBlocks
	}
	return a.LastPackInfo.Index < b.LastPackInfo.Index
}

func newPeer(version int, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return &peer{
		Peer:        p,
		rw:          rw,
		version:     version,
		id:          fmt.Sprintf("%x", p.ID().Bytes()[:8]),
		knownTxs:    mapset.NewSet(),
		knownEvents: mapset.NewSet(),
		queuedTxs:   make(chan []*types.Transaction, maxQueuedTxs),
		queuedProps: make(chan inter.Events, maxQueuedProps),
		queuedAnns:  make(chan hash.Events, maxQueuedAnns),
		term:        make(chan struct{}),
	}
}

// broadcast is a write loop that multiplexes event propagations, announcements
// and transaction broadcasts into the remote peer. The goal is to have an async
// writer that does not lock up node internals.
func (p *peer) broadcast() {
	for {
		select {
		case txs := <-p.queuedTxs:
			if err := p.SendTransactions(txs); err != nil {
				return
			}
			p.Log().Trace("Broadcast transactions", "count", len(txs))

		case events := <-p.queuedProps:
			if err := p.SendEvents(events); err != nil {
				return
			}
			p.Log().Trace("Broadcast events", "count", len(events))

		case ids := <-p.queuedAnns:
			if err := p.SendNewEventHashes(ids); err != nil {
				return
			}
			p.Log().Trace("Broadcast event hashes", "count", len(ids))

		case <-p.term:
			return
		}
	}
}

// close signals the broadcast goroutine to terminate.
func (p *peer) close() {
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

// AsyncSendTransactions queues list of transactions propagation to a remote
// peer. If the peer's broadcast queue is full, the event is silently dropped.
func (p *peer) AsyncSendTransactions(txs []*types.Transaction) {
	select {
	case p.queuedTxs <- txs:
		// Mark all the transactions as known, but ensure we don't overflow our limits
		for _, tx := range txs {
			p.knownTxs.Add(tx.Hash())
		}
		for p.knownTxs.Cardinality() >= maxKnownTxs {
			p.knownTxs.Pop()
		}
	default:
		p.Log().Debug("Dropping transaction propagation", "count", len(txs))
	}
}

// SendNewEventHashes announces the availability of a number of events through
// a hash notification.
func (p *peer) SendNewEventHashes(hashes []hash.Event) error {
	// Mark all the event hashes as known, but ensure we don't overflow our limits
	for _, hash := range hashes {
		p.knownEvents.Add(hash)
	}
	for p.knownEvents.Cardinality() >= maxKnownEvents {
		p.knownEvents.Pop()
	}
	return p2p.Send(p.rw, NewEventHashesMsg, hashes)
}

// AsyncSendNewEventHash queues the availability of a event for propagation to a
// remote peer. If the peer's broadcast queue is full, the event is silently
// dropped.
func (p *peer) AsyncSendNewEventHashes(ids hash.Events) {
	select {
	case p.queuedAnns <- ids:
		// Mark all the event hash as known, but ensure we don't overflow our limits
		for _, id := range ids {
			p.knownEvents.Add(id)
		}
		for p.knownEvents.Cardinality() >= maxKnownEvents {
			p.knownEvents.Pop()
		}
	default:
		p.Log().Debug("Dropping event announcement", "count", len(ids))
	}
}

// SendNewEvent propagates an entire event to a remote peer.
func (p *peer) SendEvents(events inter.Events) error {
	// Mark all the event hash as known, but ensure we don't overflow our limits
	for _, event := range events {
		p.knownEvents.Add(event.Hash())
		for p.knownEvents.Cardinality() >= maxKnownEvents {
			p.knownEvents.Pop()
		}
	}
	return p2p.Send(p.rw, EventsMsg, events)
}

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

func (p *peer) SendPackInfosRLP(packInfos *packInfosDataRLP) error {
	return p2p.Send(p.rw, PackInfosMsg, packInfos)
}

func (p *peer) SendPack(pack *packData) error {
	return p2p.Send(p.rw, PackMsg, pack)
}

// AsyncSendEvents queues an entire event for propagation to a remote peer. If
// the peer's broadcast queue is full, the event is silently dropped.
func (p *peer) AsyncSendEvents(events inter.Events) {
	select {
	case p.queuedProps <- events:
		// Mark all the event hash as known, but ensure we don't overflow our limits
		for _, event := range events {
			p.knownEvents.Add(event.Hash())
		}
		for p.knownEvents.Cardinality() >= maxKnownEvents {
			p.knownEvents.Pop()
		}
	default:
		p.Log().Debug("Dropping event propagation", "count", len(events))
	}
}

// SendEventHeaders sends a batch of event headers to the remote peer.
/*func (p *peer) SendEventHeaders(headers []*EvmHeader) error {
	return p2p.Send(p.rw, EventHeadersMsg, headers)
}*/

/*// RequestOneHeader is a wrapper around the header query functions to fetch a
// single header. It is used solely by the fetcher.
func (p *peer) RequestOneHeader(hash common.Hash) error {
	p.Log().Debug("Fetching single header", "hash", hash)
	return p2p.Send(p.rw, GetEventHeadersMsg, &getEventHeadersData{Origin: hashOrNumber{Hash: hash}, Amount: uint64(1), Skip: uint64(0), Reverse: false})
}

// RequestHeadersByHash fetches a batch of events' headers corresponding to the
// specified header query, based on the hash of an origin event.
func (p *peer) RequestHeadersByHash(origin common.Hash, amount int, skip int, reverse bool) error {
	p.Log().Debug("Fetching batch of headers", "count", amount, "fromhash", origin, "skip", skip, "reverse", reverse)
	return p2p.Send(p.rw, GetEventHeadersMsg, &getEventHeadersData{Origin: hashOrNumber{Hash: origin}, Amount: uint64(amount), Skip: uint64(skip), Reverse: reverse})
}

// RequestHeadersByNumber fetches a batch of events' headers corresponding to the
// specified header query, based on the number of an origin event.
func (p *peer) RequestHeadersByNumber(origin uint64, amount int, skip int, reverse bool) error {
	p.Log().Debug("Fetching batch of headers", "count", amount, "fromnum", origin, "skip", skip, "reverse", reverse)
	return p2p.Send(p.rw, GetEventHeadersMsg, &getEventHeadersData{Origin: hashOrNumber{Number: origin}, Amount: uint64(amount), Skip: uint64(skip), Reverse: reverse})
}*/

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

func (p *peer) RequestPackInfos(epoch idx.Epoch, indexes []idx.Pack) error {
	return p2p.Send(p.rw, GetPackInfosMsg, getPackInfosData{
		Epoch:   epoch,
		Indexes: indexes,
	})
}

func (p *peer) RequestPack(epoch idx.Epoch, index idx.Pack) error {
	return p2p.Send(p.rw, GetPackMsg, getPackData{
		Epoch: epoch,
		Index: index,
	})
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
		fmt.Sprintf("lachesis/%2d", p.version),
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
	go p.broadcast()

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
