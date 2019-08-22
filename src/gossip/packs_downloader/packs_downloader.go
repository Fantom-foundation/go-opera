package packs_downloader

import (
	"github.com/Fantom-foundation/go-lachesis/src/gossip/fetcher"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/ethereum/go-ethereum/log"
	"sync"
)

const (
	maxPeers = 6 // max peers to download packs from
)

// PacksDownloader is responsible for accumulating pack announcements from various peers
// and scheduling them for retrieval.
type PacksDownloader struct {
	// Callbacks
	dropPeer         dropPeerFn
	fetcher          *fetcher.Fetcher
	onlyNotConnected onlyNotConnectedFn

	// State
	peers map[string]*PeerPacksDownloader

	peersMu *sync.RWMutex
}

// New creates a packs fetcher to retrieve events based on pack announcements.
func New(fetcher *fetcher.Fetcher, onlyNotConnected onlyNotConnectedFn, dropPeer dropPeerFn) *PacksDownloader {
	return &PacksDownloader{
		fetcher:          fetcher,
		onlyNotConnected: onlyNotConnected,
		dropPeer:         dropPeer,
		peers:            make(map[string]*PeerPacksDownloader),
		peersMu:          new(sync.RWMutex),
	}
}

type Peer struct {
	Id    string
	Epoch idx.SuperFrame

	RequestPackInfos packInfoRequesterFn
	RequestPack      packRequesterFn
}

// RegisterPeer injects a new download peer into the set of block source to be
// used for fetching hashes and blocks from.
func (d *PacksDownloader) RegisterPeer(peer Peer, myEpoch idx.SuperFrame) error {
	if peer.Epoch < myEpoch {
		// this peer is useless for syncing
		return d.UnregisterPeer(peer.Id)
	}

	d.peersMu.Lock()
	defer d.peersMu.Unlock()

	if d.peers[peer.Id] != nil || len(d.peers) >= maxPeers {
		return nil
	}

	log.Trace("Registering sync peer", "peer", peer, "epoch", myEpoch)
	d.peers[peer.Id] = newPeer(peer, myEpoch, d.fetcher, d.onlyNotConnected, d.dropPeer)
	d.peers[peer.Id].Start()

	return nil
}

func (d *PacksDownloader) OnNewEpoch(myEpoch idx.SuperFrame, peerEpoch func(string) idx.SuperFrame) {
	newPeers := make(map[string]*PeerPacksDownloader)

	for peerId, peerDwnld := range d.peers {
		peerDwnld.Stop()

		if peerEpoch(peerId) >= myEpoch {
			// allocate new peer for the new epoch
			newPeerDwnld := newPeer(peerDwnld.peer, myEpoch, d.fetcher, d.onlyNotConnected, d.dropPeer)
			newPeerDwnld.Start()
			newPeers[peerId] = newPeerDwnld
		} else {
			log.Trace("UnRegistering sync peer", "peer", peerId)
		}
	}
	// wipe out old downloading state from prev. epoch
	d.peers = newPeers
}

func (d *PacksDownloader) Peer(peer string) *PeerPacksDownloader {
	d.peersMu.RLock()
	defer d.peersMu.RUnlock()

	return d.peers[peer]
}

func (d *PacksDownloader) PeersNum() int {
	d.peersMu.RLock()
	defer d.peersMu.RUnlock()

	return len(d.peers)
}

// UnregisterPeer removes a peer from the known list, preventing any action from
// the specified peer. An effort is also made to return any pending fetches into
// the queue.
func (d *PacksDownloader) UnregisterPeer(peer string) error {
	d.peersMu.Lock()
	defer d.peersMu.Unlock()

	if d.peers[peer] == nil {
		return nil
	}

	log.Trace("UnRegistering sync peer", "peer", peer)
	d.peers[peer].Stop()
	delete(d.peers, peer)
	return nil
}

// Terminate interrupts the downloader, canceling all pending operations.
// The downloader cannot be reused after calling Terminate.
func (d *PacksDownloader) Terminate() {
	d.peersMu.Lock()
	defer d.peersMu.Unlock()

	for _, peerDownloader := range d.peers {
		peerDownloader.Stop()
	}
	d.peers = make(map[string]*PeerPacksDownloader)
}
