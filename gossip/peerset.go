// Copyright 2020 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package gossip

import (
	"errors"
	"sync"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth/protocols/snap"
	"github.com/ethereum/go-ethereum/p2p"
)

var (
	// errPeerSetClosed is returned if a peer is attempted to be added or removed
	// from the peer set after it has been terminated.
	errPeerSetClosed = errors.New("peerset closed")

	// errPeerAlreadyRegistered is returned if a peer is attempted to be added
	// to the peer set, but one with the same id already exists.
	errPeerAlreadyRegistered = errors.New("peer already registered")

	// errPeerNotRegistered is returned if a peer is attempted to be removed from
	// a peer set, but no peer with the given id exists.
	errPeerNotRegistered = errors.New("peer not registered")

	// errSnapWithoutOpera is returned if a peer attempts to connect only on the
	// snap protocol without advertizing the opera main protocol.
	errSnapWithoutOpera = errors.New("peer connected on snap without compatible opera support")
)

// peerSet represents the collection of active peers currently participating in
// the `eth` protocol, with or without the `snap` extension.
type peerSet struct {
	peers     map[string]*peer // Peers connected on the `eth` protocol
	snapPeers int              // Number of `snap` compatible peers for connection prioritization

	snapWait map[string]chan *snap.Peer // Peers connected on `eth` waiting for their snap extension
	snapPend map[string]*snap.Peer      // Peers connected on the `snap` protocol, but not yet on `eth`

	lock   sync.RWMutex
	closed bool
}

// newPeerSet creates a new peer set to track the active participants.
func newPeerSet() *peerSet {
	return &peerSet{
		peers:    make(map[string]*peer),
		snapWait: make(map[string]chan *snap.Peer),
		snapPend: make(map[string]*snap.Peer),
	}
}

// RegisterSnapExtension unblocks an already connected `eth` peer waiting for its
// `snap` extension, or if no such peer exists, tracks the extension for the time
// being until the `eth` main protocol starts looking for it.
func (ps *peerSet) RegisterSnapExtension(peer *snap.Peer) error {
	// Reject the peer if it is not eligible for a snap protocol
	if !eligibleForSnap(peer.Peer) {
		return errSnapWithoutOpera
	}
	// Ensure nobody can double connect
	ps.lock.Lock()
	defer ps.lock.Unlock()

	id := peer.ID()
	if _, ok := ps.peers[id]; ok {
		return errPeerAlreadyRegistered // avoid connections with the same id as existing ones
	}
	if _, ok := ps.snapPend[id]; ok {
		return errPeerAlreadyRegistered // avoid connections with the same id as pending ones
	}
	// Inject the peer into an `eth` counterpart is available, otherwise save for later
	if wait, ok := ps.snapWait[id]; ok {
		delete(ps.snapWait, id)
		wait <- peer
		return nil
	}
	ps.snapPend[id] = peer
	return nil
}

// WaitSnapExtension blocks until all satellite protocols are connected and tracked
// by the peerset.
func (ps *peerSet) WaitSnapExtension(p *peer) (*snap.Peer, error) {
	// If the peer is not eligible for a snap protocol`, don't wait
	if !eligibleForSnap(p.Peer) {
		return nil, nil
	}
	// Ensure nobody can double connect
	ps.lock.Lock()

	id := p.id
	if _, ok := ps.peers[id]; ok {
		ps.lock.Unlock()
		return nil, errPeerAlreadyRegistered // avoid connections with the same id as existing ones
	}
	if _, ok := ps.snapWait[id]; ok {
		ps.lock.Unlock()
		return nil, errPeerAlreadyRegistered // avoid connections with the same id as pending ones
	}
	// If `snap` already connected, retrieve the peer from the pending set
	if snap, ok := ps.snapPend[id]; ok {
		delete(ps.snapPend, id)

		ps.lock.Unlock()
		return snap, nil
	}
	// Otherwise wait for `snap` to connect concurrently
	wait := make(chan *snap.Peer)
	ps.snapWait[id] = wait
	ps.lock.Unlock()

	return <-wait, nil
}

// RegisterPeer injects a new `eth` peer into the working set, or returns an error
// if the peer is already known.
func (ps *peerSet) RegisterPeer(p *peer, ext *snap.Peer) error {
	// Start tracking the new peer
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if ps.closed {
		return errPeerSetClosed
	}

	id := p.id
	if _, ok := ps.peers[id]; ok {
		return errPeerAlreadyRegistered
	}

	if ext != nil {
		p.snapExt = &snapPeer{ext}
		ps.snapPeers++
	}

	ps.peers[id] = p
	return nil
}

// UnregisterPeer removes a remote peer from the active set, disabling any further
// actions to/from that particular entity.
func (ps *peerSet) UnregisterPeer(id string) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	peer, ok := ps.peers[id]
	if !ok {
		return errPeerNotRegistered
	}
	delete(ps.peers, id)
	if peer.snapExt != nil {
		ps.snapPeers--
	}
	return nil
}

// Peer retrieves the registered peer with the given id.
func (ps *peerSet) Peer(id string) *peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return ps.peers[id]
}

func (ps *peerSet) UselessNum() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	num := 0
	for _, p := range ps.peers {
		if p.Useless() {
			num++
		}
	}
	return num
}

// PeersWithoutEvent retrieves a list of peers that do not have a given event in
// their set of known hashes so it might be propagated to them.
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

// PeersWithoutTx retrieves a list of peers that do not have a given
// transaction in their set of known hashes.
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

// List returns array of peers in the set.
func (ps *peerSet) List() []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		list = append(list, p)
	}
	return list
}

// Len returns if the current number of `eth` peers in the set. Since the `snap`
// peers are tied to the existence of an `eth` connection, that will always be a
// subset of `eth`.
func (ps *peerSet) Len() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return len(ps.peers)
}

// SnapLen returns if the current number of `snap` peers in the set.
func (ps *peerSet) SnapLen() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return ps.snapPeers
}

// Close disconnects all peers.
func (ps *peerSet) Close() {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for _, p := range ps.peers {
		p.Disconnect(p2p.DiscQuitting)
	}
	ps.closed = true
}
