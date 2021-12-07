package snapleecher

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/eth/protocols/snap"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
)

var (
	MaxStateFetch = 384 // Amount of node state values to allow fetching per request
)

var (
	errTimeout          = errors.New("timeout")
	errCancelStateFetch = errors.New("state data download canceled (requested)")
	errCanceled         = errors.New("syncing canceled (requested)")
)

type Leecher struct {
	peers *peerSet // Set of active peers from which download can proceed
	// Callbacks
	dropPeer peerDropFn // Drops a peer for misbehaving

	stateDB    ethdb.Database  // Database to state sync into (and deduplicate via)
	stateBloom *trie.SyncBloom // Bloom filter for fast trie node and contract code existence checks
	SnapSyncer *snap.Syncer    // TODO(karalabe): make private! hack for now

	stateSyncStart chan *stateSync
	trackStateReq  chan *stateReq
	stateCh        chan dataPack // Channel receiving inbound node state data

	// Statistics
	syncStatsState stateSyncStats
	syncStatsLock  sync.RWMutex // Lock protecting the sync stats fields

	// Cancellation and termination
	cancelPeer string         // Identifier of the peer currently being used as the master (cancel on drop)
	cancelCh   chan struct{}  // Channel to cancel mid-flight syncs
	cancelLock sync.RWMutex   // Lock to protect the cancel channel and peer in delivers
	cancelWg   sync.WaitGroup // Make sure all fetcher goroutines have exited.

	quitCh chan struct{} // Quit channel to signal termination
}

// New creates a new downloader to fetch hashes and blocks from remote peers.
func New(stateDb ethdb.Database, stateBloom *trie.SyncBloom, dropPeer peerDropFn) *Leecher {
	d := &Leecher{
		stateDB:        stateDb,
		stateBloom:     stateBloom,
		dropPeer:       dropPeer,
		peers:          newPeerSet(),
		quitCh:         make(chan struct{}),
		stateCh:        make(chan dataPack),
		SnapSyncer:     snap.NewSyncer(stateDb),
		stateSyncStart: make(chan *stateSync),
		syncStatsState: stateSyncStats{
			processed: rawdb.ReadFastTrieProgress(stateDb),
		},
		trackStateReq: make(chan *stateReq),
	}
	go d.stateFetcher()
	return d
}

// cancel aborts all of the operations and resets the queue. However, cancel does
// not wait for the running download goroutines to finish. This method should be
// used when cancelling the downloads from inside the downloader.
func (d *Leecher) cancel() {
	// Close the current cancel channel
	d.cancelLock.Lock()
	defer d.cancelLock.Unlock()

	if d.cancelCh != nil {
		select {
		case <-d.cancelCh:
			// Channel was already closed
		default:
			close(d.cancelCh)
		}
	}
}

// DeliverSnapPacket is invoked from a peer's message handler when it transmits a
// data packet for the local node to consume.
func (d *Leecher) DeliverSnapPacket(peer *snap.Peer, packet snap.Packet) error {
	switch packet := packet.(type) {
	case *snap.AccountRangePacket:
		hashes, accounts, err := packet.Unpack()
		if err != nil {
			return err
		}
		return d.SnapSyncer.OnAccounts(peer, packet.ID, hashes, accounts, packet.Proof)

	case *snap.StorageRangesPacket:
		hashset, slotset := packet.Unpack()
		return d.SnapSyncer.OnStorage(peer, packet.ID, hashset, slotset, packet.Proof)

	case *snap.ByteCodesPacket:
		return d.SnapSyncer.OnByteCodes(peer, packet.ID, packet.Codes)

	case *snap.TrieNodesPacket:
		return d.SnapSyncer.OnTrieNodes(peer, packet.ID, packet.Nodes)

	default:
		return fmt.Errorf("unexpected snap packet type: %T", packet)
	}
}

// Progress retrieves the synchronisation boundaries, specifically the origin
// block where synchronisation started at (may have failed/suspended); the block
// or header sync is currently at; and the latest known block which the sync targets.
//
// In addition, during the state download phase of fast synchronisation the number
// of processed and the total number of known states are also returned. Otherwise
// these are zero.
func (d *Leecher) Progress() ethereum.SyncProgress {
	// Lock the current stats and return the progress
	d.syncStatsLock.RLock()
	defer d.syncStatsLock.RUnlock()

	return ethereum.SyncProgress{
		PulledStates: d.syncStatsState.processed,
		KnownStates:  d.syncStatsState.processed + d.syncStatsState.pending,
	}
}
