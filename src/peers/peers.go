package peers

import (
	"sort"
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// PubKeyPeers map of peers sorted by public key
type PubKeyPeers map[string]*Peer

// IDPeers map of peers sorted by ID
type IDPeers map[uint64]*Peer

// AddressPeers maps address to peer
type AddressPeers map[common.Address]*Peer

type NetAddrPeers map[string]*Peer

// Listener for listening for new peers joining
type Listener func(*Peer)

// Peers struct for all known peers for this node
type Peers struct {
	sync.RWMutex
	Sorted    []*Peer
	ByPubKey  PubKeyPeers
	ByID      IDPeers
	ByAddress AddressPeers
	ByNetAddr NetAddrPeers
	Listeners []Listener
}

/* Constructors */

// NewPeers creates a new peers struct
func NewPeers() *Peers {
	return &Peers{
		ByPubKey:  make(PubKeyPeers),
		ByID:      make(IDPeers),
		ByAddress: make(AddressPeers),
		ByNetAddr: make(NetAddrPeers),
	}
}

// NewPeersFromSlice create a new peers struct from a subset of peers
func NewPeersFromSlice(source []*Peer) *Peers {
	peers := NewPeers()

	for _, peer := range source {
		peers.addPeerRaw(peer)
	}

	peers.internalSort()

	return peers
}

/* Add Methods */

// Add a peer without sorting the set.
// Useful for adding a bunch of peers at the same time
// This method is private and is not protected by mutex.
// Handle with care
func (p *Peers) addPeerRaw(peer *Peer) {
	if peer.ID == 0 {
		if err := peer.computeID(); err != nil {
			panic(err)
		}
	}

	p.ByPubKey[peer.PubKeyHex] = peer
	p.ByID[peer.ID] = peer
	p.ByAddress[peer.Address()] = peer
	p.ByNetAddr[peer.NetAddr] = peer
}

// AddPeer adds a peer to the peers struct
func (p *Peers) AddPeer(peer *Peer) {
	p.Lock()
	p.addPeerRaw(peer)
	p.internalSort()
	p.Unlock()
	p.EmitNewPeer(peer)
}

func (p *Peers) internalSort() {
	res := []*Peer{}

	for _, p := range p.ByPubKey {
		res = append(res, p)
	}

	sort.Sort(ByID(res))

	p.Sorted = res
}

/* Remove Methods */

// RemovePeer removes a peer from the peers struct
func (p *Peers) RemovePeer(peer *Peer) {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.ByPubKey[peer.PubKeyHex]; !ok {
		return
	}

	delete(p.ByPubKey, peer.PubKeyHex)
	delete(p.ByID, peer.ID)
	delete(p.ByAddress, peer.Address())
	delete(p.ByNetAddr, peer.NetAddr)

	p.internalSort()
}

// RemovePeerByPubKey removes a peer by their public key
func (p *Peers) RemovePeerByPubKey(pubKey string) {
	p.RemovePeer(p.ByPubKey[pubKey])
}

// RemovePeerByID removes a peer based on their ID
func (p *Peers) RemovePeerByID(id uint64) {
	p.RemovePeer(p.ByID[id])
}

/* ToSlice Methods */

// ToPeerSlice returns a slice of peers sorted
func (p *Peers) ToPeerSlice() []*Peer {
	return p.Sorted
}

// ToPeerByUsedSlice sorted peers list
func (p *Peers) ToPeerByUsedSlice() []*Peer {
	res := []*Peer{}

	for _, p := range p.ByPubKey {
		res = append(res, p)
	}

	sort.Sort(ByUsed(res))
	return res
}

// ToPubKeySlice peers struct by public key
func (p *Peers) ToPubKeySlice() []string {
	p.RLock()
	defer p.RUnlock()

	res := []string{}

	for _, peer := range p.Sorted {
		res = append(res, peer.PubKeyHex)
	}

	return res
}

// ToIDSlice peers struct by ID
func (p *Peers) ToIDSlice() []uint64 {
	p.RLock()
	defer p.RUnlock()

	res := []uint64{}

	for _, peer := range p.Sorted {
		res = append(res, peer.ID)
	}

	return res
}

/* EventListener */

// OnNewPeer on new peer joined event trigger listener
func (p *Peers) OnNewPeer(cb func(*Peer)) {
	p.Listeners = append(p.Listeners, cb)
}

// EmitNewPeer emits an event for all listeners as soon as a peer joins
func (p *Peers) EmitNewPeer(peer *Peer) {
	for _, listener := range p.Listeners {
		listener(peer)
	}
}

/* Utilities */

// Len returns the length of peers
func (p *Peers) Len() int {
	p.RLock()
	defer p.RUnlock()

	return len(p.ByPubKey)
}

func (p *Peers) ReadByPubKey(key string) (Peer, bool) {
	p.RLock()
	defer p.RUnlock()
	peer, ok := p.ByPubKey[key]
	return *peer, ok
}

func (p *Peers) ReadByID(key uint64) (Peer, bool) {
	p.RLock()
	defer p.RUnlock()
	peer, ok := p.ByID[key]
	return *peer, ok
}

func (p *Peers) ReadByAddress(key common.Address) (Peer, bool) {
	p.RLock()
	defer p.RUnlock()
	peer, ok := p.ByAddress[key]
	return *peer, ok
}

func (p *Peers) ReadByNetAddr(key string) (Peer, bool) {
	p.RLock()
	defer p.RUnlock()
	peer, ok := p.ByNetAddr[key]
	return *peer, ok
}

// ByPubHex implements sort.Interface for Peers based on
// the PubKeyHex field.
type ByPubHex []*Peer

func (a ByPubHex) Len() int      { return len(a) }
func (a ByPubHex) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByPubHex) Less(i, j int) bool {
	ai := a[i].PubKeyHex
	aj := a[j].PubKeyHex
	return ai < aj
}

// ByID sorted by ID peers list
type ByID []*Peer

func (a ByID) Len() int      { return len(a) }
func (a ByID) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByID) Less(i, j int) bool {
	ai := a[i].ID
	aj := a[j].ID
	return ai < aj
}

// ByUsed TODO
type ByUsed []*Peer

func (a ByUsed) Len() int      { return len(a) }
func (a ByUsed) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByUsed) Less(i, j int) bool {
	ai := a[i].Used
	aj := a[j].Used
	return ai > aj
}
