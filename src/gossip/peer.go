package gossip

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/mclock"
	"github.com/ethereum/go-ethereum/les/flowcontrol"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	errClosed            = errors.New("peer set is closed")
	errAlreadyRegistered = errors.New("peer is already registered")
	errNotRegistered     = errors.New("peer is not registered")
)

const (
	maxRequestErrors  = 20 // number of invalid requests tolerated (makes the protocol less brittle but still avoids spam)
	maxResponseErrors = 50 // number of invalid responses tolerated (makes the protocol less brittle but still avoids spam)
)

// capacity limitation for parameter updates
const (
	allowedUpdateBytes = 100000                // initial/maximum allowed update size
	allowedUpdateRate  = time.Millisecond * 10 // time constant for recharging one byte of allowance
)

const (
	freezeTimeBase    = time.Millisecond * 700 // fixed component of client freeze time
	freezeTimeRandom  = time.Millisecond * 600 // random component of client freeze time
	freezeCheckPeriod = time.Millisecond * 100 // buffer value recheck period after initial freeze time has elapsed
)

// if the total encoded size of a sent transaction batch is over txSizeCostLimit
// per transaction then the request cost is calculated as proportional to the
// encoded size instead of the transaction count
const txSizeCostLimit = 0x4000

type peer struct {
	*p2p.Peer
	rw p2p.MsgReadWriter

	version int    // Protocol version negotiated
	network uint64 // Network ID being on

	// Checkpoint relative fields
	checkpoint       params.TrustedCheckpoint
	checkpointNumber uint64

	id string

	lock sync.RWMutex

	sendQueue *execQueue

	errCh chan error
	// responseLock ensures that responses are queued in the same order as
	// RequestProcessed is called
	responseLock  sync.Mutex
	responseCount uint64
	invalidCount  uint32

	poolEntry      *poolEntry
	responseErrors int
	updateCounter  uint64
	updateTime     mclock.AbsTime
	frozen         uint32 // 1 if client is in frozen state

	fcClient *flowcontrol.ClientNode // nil if the peer is server only
	fcServer *flowcontrol.ServerNode // nil if the peer is client only
}

func newPeer(version int, network uint64, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return &peer{
		Peer:    p,
		rw:      rw,
		version: version,
		network: network,
		id:      fmt.Sprintf("%x", p.ID().Bytes()),
		errCh:   make(chan error, 1),
	}
}

// rejectUpdate returns true if a parameter update has to be rejected because
// the size and/or rate of updates exceed the capacity limitation
func (p *peer) rejectUpdate(size uint64) bool {
	now := mclock.Now()
	if p.updateCounter == 0 {
		p.updateTime = now
	} else {
		dt := now - p.updateTime
		r := uint64(dt / mclock.AbsTime(allowedUpdateRate))
		if p.updateCounter > r {
			p.updateCounter -= r
			p.updateTime += mclock.AbsTime(allowedUpdateRate * time.Duration(r))
		} else {
			p.updateCounter = 0
			p.updateTime = now
		}
	}
	p.updateCounter += size
	return p.updateCounter > allowedUpdateBytes
}

// freezeClient temporarily puts the client in a frozen state which means all
// unprocessed and subsequent requests are dropped. Unfreezing happens automatically
// after a short time if the client's buffer value is at least in the slightly positive
// region. The client is also notified about being frozen/unfrozen with a Stop/Resume
// message.
func (p *peer) freezeClient() {
	if atomic.SwapUint32(&p.frozen, 1) != 0 {
		return
	}
	go func() {
		p.SendStop()
		time.Sleep(freezeTimeBase + time.Duration(rand.Int63n(int64(freezeTimeRandom))))
		for {
			bufValue, bufLimit := p.fcClient.BufferStatus()
			if bufLimit == 0 {
				return
			}
			if bufValue <= bufLimit/8 {
				time.Sleep(freezeCheckPeriod)
			} else {
				atomic.StoreUint32(&p.frozen, 0)
				p.SendResume(bufValue)
				break
			}
		}
	}()
}

// freezeServer processes Stop/Resume messages from the given server
func (p *peer) freezeServer(frozen bool) {
	var f uint32
	if frozen {
		f = 1
	}
	if atomic.SwapUint32(&p.frozen, f) != f && frozen {
		p.sendQueue.clear()
	}
}

// isFrozen returns true if the client is frozen or the server has put our
// client in frozen state
func (p *peer) isFrozen() bool {
	return atomic.LoadUint32(&p.frozen) != 0
}

func (p *peer) canQueue() bool {
	return p.sendQueue.canQueue() && !p.isFrozen()
}

func (p *peer) queueSend(f func()) {
	p.sendQueue.queue(f)
}

// waitBefore implements distPeer interface
func (p *peer) waitBefore(maxCost uint64) (time.Duration, float64) {
	return p.fcServer.CanSend(maxCost)
}

func sendRequest(w p2p.MsgWriter, msgcode, reqID, cost uint64, data interface{}) error {
	type req struct {
		ReqID uint64
		Data  interface{}
	}
	return p2p.Send(w, msgcode, req{reqID, data})
}

// reply struct represents a reply with the actual data already RLP encoded and
// only the bv (buffer value) missing. This allows the serving mechanism to
// calculate the bv value which depends on the data size before sending the reply.
type reply struct {
	w              p2p.MsgWriter
	msgcode, reqID uint64
	data           rlp.RawValue
}

// send sends the reply with the calculated buffer value
func (r *reply) send(bv uint64) error {
	type resp struct {
		ReqID, BV uint64
		Data      rlp.RawValue
	}
	return p2p.Send(r.w, r.msgcode, resp{r.reqID, bv, r.data})
}

// size returns the RLP encoded size of the message data
func (r *reply) size() uint32 {
	return uint32(len(r.data))
}

// SendStop notifies the client about being in frozen state
func (p *peer) SendStop() error {
	return p2p.Send(p.rw, StopMsg, struct{}{})
}

// SendResume notifies the client about getting out of frozen state
func (p *peer) SendResume(bv uint64) error {
	return p2p.Send(p.rw, ResumeMsg, bv)
}

// ReplyCode creates a reply with a batch of arbitrary internal data, corresponding to the
// hashes requested.
func (p *peer) ReplyCode(reqID uint64, codes [][]byte) *reply {
	data, _ := rlp.EncodeToBytes(codes)
	return &reply{p.rw, CodeMsg, reqID, data}
}

type keyValueEntry struct {
	Key   string
	Value rlp.RawValue
}
type keyValueList []keyValueEntry
type keyValueMap map[string]rlp.RawValue

func (l keyValueList) add(key string, val interface{}) keyValueList {
	var entry keyValueEntry
	entry.Key = key
	if val == nil {
		val = uint64(0)
	}
	enc, err := rlp.EncodeToBytes(val)
	if err == nil {
		entry.Value = enc
	}
	return append(l, entry)
}

func (l keyValueList) decode() (keyValueMap, uint64) {
	m := make(keyValueMap)
	var size uint64
	for _, entry := range l {
		m[entry.Key] = entry.Value
		size += uint64(len(entry.Key)) + uint64(len(entry.Value)) + 8
	}
	return m, size
}

func (m keyValueMap) get(key string, val interface{}) error {
	enc, ok := m[key]
	if !ok {
		return errResp(ErrMissingKey, "%s", key)
	}
	if val == nil {
		return nil
	}
	return rlp.DecodeBytes(enc, val)
}

func (p *peer) sendReceiveHandshake(sendList keyValueList) (keyValueList, error) {
	// Send out own handshake in a new thread
	errc := make(chan error, 1)
	go func() {
		errc <- p2p.Send(p.rw, StatusMsg, sendList)
	}()
	// In the mean time retrieve the remote status message
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return nil, err
	}
	if msg.Code != StatusMsg {
		return nil, errResp(ErrNoStatusMsg, "first msg has code %x (!= %x)", msg.Code, StatusMsg)
	}
	if msg.Size > ProtocolMaxMsgSize {
		return nil, errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	// Decode the handshake
	var recvList keyValueList
	if err := msg.Decode(&recvList); err != nil {
		return nil, errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	if err := <-errc; err != nil {
		return nil, err
	}
	return recvList, nil
}

// Handshake executes the lachesis protocol handshake, negotiating version number,
// network IDs and genesis blocks.
func (p *peer) Handshake(genesis common.Hash, server *Service) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	var send keyValueList
	send = send.add("protocolVersion", uint64(p.version))
	send = send.add("networkId", p.network)
	send = send.add("genesisHash", genesis)
	if server != nil {
		// TODO: handshake data from server
	}

	recvList, err := p.sendReceiveHandshake(send)
	if err != nil {
		return err
	}
	recv, size := recvList.decode()
	if p.rejectUpdate(size) {
		return errResp(ErrRequestRejected, "")
	}

	var (
		rVersion uint64
		rNetwork uint64
		rGenesis common.Hash
	)

	if err := recv.get("protocolVersion", &rVersion); err != nil {
		return err
	}
	if err := recv.get("networkId", &rNetwork); err != nil {
		return err
	}
	if err := recv.get("genesisHash", &rGenesis); err != nil {
		return err
	}

	if rGenesis != genesis {
		return errResp(ErrGenesisBlockMismatch, "%x (!= %x)", rGenesis[:8], genesis[:8])
	}
	if rNetwork != p.network {
		return errResp(ErrNetworkIdMismatch, "%d (!= %d)", rNetwork, p.network)
	}
	if int(rVersion) != p.version {
		return errResp(ErrProtocolVersionMismatch, "%d (!= %d)", rVersion, p.version)
	}

	return nil
}

// String implements fmt.Stringer.
func (p *peer) String() string {
	return fmt.Sprintf("Peer %s [%s]", p.id,
		fmt.Sprintf("lachesis/%d", p.version),
	)
}

// peerSetNotify is a callback interface to notify services about added or
// removed peers
type peerSetNotify interface {
	registerPeer(*peer)
	unregisterPeer(*peer)
}

// peerSet represents the collection of active peers currently participating in
// the Light Ethereum sub-protocol.
type peerSet struct {
	peers      map[string]*peer
	lock       sync.RWMutex
	notifyList []peerSetNotify
	closed     bool
}

// newPeerSet creates a new peer set to track the active participants.
func newPeerSet() *peerSet {
	return &peerSet{
		peers: make(map[string]*peer),
	}
}

// notify adds a service to be notified about added or removed peers
func (ps *peerSet) notify(n peerSetNotify) {
	ps.lock.Lock()
	ps.notifyList = append(ps.notifyList, n)
	peers := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		peers = append(peers, p)
	}
	ps.lock.Unlock()

	for _, p := range peers {
		n.registerPeer(p)
	}
}

// Register injects a new peer into the working set, or returns an error if the
// peer is already known.
func (ps *peerSet) Register(p *peer) error {
	ps.lock.Lock()
	if ps.closed {
		ps.lock.Unlock()
		return errClosed
	}
	if _, ok := ps.peers[p.id]; ok {
		ps.lock.Unlock()
		return errAlreadyRegistered
	}
	ps.peers[p.id] = p
	p.sendQueue = newExecQueue(100)
	peers := make([]peerSetNotify, len(ps.notifyList))
	copy(peers, ps.notifyList)
	ps.lock.Unlock()

	for _, n := range peers {
		n.registerPeer(p)
	}
	return nil
}

// Unregister removes a remote peer from the active set, disabling any further
// actions to/from that particular entity. It also initiates disconnection at the networking layer.
func (ps *peerSet) Unregister(id string) error {
	ps.lock.Lock()
	if p, ok := ps.peers[id]; !ok {
		ps.lock.Unlock()
		return errNotRegistered
	} else {
		delete(ps.peers, id)
		peers := make([]peerSetNotify, len(ps.notifyList))
		copy(peers, ps.notifyList)
		ps.lock.Unlock()

		for _, n := range peers {
			n.unregisterPeer(p)
		}

		p.sendQueue.quit()
		p.Peer.Disconnect(p2p.DiscUselessPeer)

		return nil
	}
}

// AllPeerIDs returns a list of all registered peer IDs
func (ps *peerSet) AllPeerIDs() []string {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	res := make([]string, len(ps.peers))
	idx := 0
	for id := range ps.peers {
		res[idx] = id
		idx++
	}
	return res
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

// AllPeers returns all peers in a list
func (ps *peerSet) AllPeers() []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, len(ps.peers))
	i := 0
	for _, peer := range ps.peers {
		list[i] = peer
		i++
	}
	return list
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
