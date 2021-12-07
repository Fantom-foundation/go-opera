// Copyright 2017 The go-ethereum Authors
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

package snapleecher

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/trie"
	"golang.org/x/crypto/sha3"
)

// stateReq represents a batch of state fetch requests grouped together into
// a single data retrieval network packet.
type stateReq struct {
	nItems    uint16                    // Number of items requested for download (max is 384, so uint16 is sufficient)
	trieTasks map[common.Hash]*trieTask // Trie node download tasks to track previous attempts
	codeTasks map[common.Hash]*codeTask // Byte code download tasks to track previous attempts
	timeout   time.Duration             // Maximum round trip time for this to complete
	timer     *time.Timer               // Timer to fire when the RTT timeout expires
	peer      *peerConnection           // Peer that we're requesting from
	delivered time.Time                 // Time when the packet was delivered (independent when we process it)
	response  [][]byte                  // Response data of the peer (nil for timeouts)
	dropped   bool                      // Flag whether the peer dropped off early
}

// timedOut returns if this request timed out.
func (req *stateReq) timedOut() bool {
	return req.response == nil
}

// stateSyncStats is a collection of progress stats to report during a state trie
// sync to RPC requests as well as to display in user logs.
type stateSyncStats struct {
	processed  uint64 // Number of state entries processed
	duplicate  uint64 // Number of state entries downloaded twice
	unexpected uint64 // Number of non-requested state entries received
	pending    uint64 // Number of still pending state entries
}

// SyncState starts downloading state with the given root hash.
func (d *Leecher) SyncState(root common.Hash) *stateSync {
	// Create the state sync
	s := newStateSync(d, root)
	select {
	case d.stateSyncStart <- s:
		// If we tell the statesync to restart with a new root, we also need
		// to wait for it to actually also start -- when old requests have timed
		// out or been delivered
		<-s.started
	case <-d.quitCh:
		s.err = errCancelStateFetch
		close(s.done)
	}
	return s
}

// stateFetcher manages the active state sync and accepts requests
// on its behalf.
func (d *Leecher) stateFetcher() {
	for {
		select {
		case s := <-d.stateSyncStart:
			for next := s; next != nil; {
				next = d.runStateSync(next)
			}
		case <-d.stateCh:
			// Ignore state responses while no sync is running.
		case <-d.quitCh:
			return
		}
	}
}

// runStateSync runs a state synchronisation until it completes or another root
// hash is requested to be switched over to.
func (d *Leecher) runStateSync(s *stateSync) *stateSync {
	var (
		active   = make(map[string]*stateReq) // Currently in-flight requests
		finished []*stateReq                  // Completed or failed requests
		timeout  = make(chan *stateReq)       // Timed out active requests
	)
	log.Trace("State sync starting", "root", s.root)

	defer func() {
		// Cancel active request timers on exit. Also set peers to idle so they're
		// available for the next sync.
		for _, req := range active {
			req.timer.Stop()
			req.peer.SetNodeDataIdle(int(req.nItems), time.Now())
		}
	}()
	go s.run()
	defer s.Cancel()

	// Listen for peer departure events to cancel assigned tasks
	peerDrop := make(chan *peerConnection, 1024)
	peerSub := s.d.peers.SubscribePeerDrops(peerDrop)
	defer peerSub.Unsubscribe()

	for {
		// Enable sending of the first buffered element if there is one.
		var (
			deliverReq   *stateReq
			deliverReqCh chan *stateReq
		)
		if len(finished) > 0 {
			deliverReq = finished[0]
			deliverReqCh = s.deliver
		}

		select {
		// The stateSync lifecycle:
		case next := <-d.stateSyncStart:
			d.spindownStateSync(active, finished, timeout, peerDrop)
			return next

		case <-s.done:
			d.spindownStateSync(active, finished, timeout, peerDrop)
			return nil

		// Send the next finished request to the current sync:
		case deliverReqCh <- deliverReq:
			// Shift out the first request, but also set the emptied slot to nil for GC
			copy(finished, finished[1:])
			finished[len(finished)-1] = nil
			finished = finished[:len(finished)-1]

		// Handle incoming state packs:
		case pack := <-d.stateCh:
			// Discard any data not requested (or previously timed out)
			req := active[pack.PeerId()]
			if req == nil {
				log.Debug("Unrequested node data", "peer", pack.PeerId(), "len", pack.Items())
				continue
			}
			// Finalize the request and queue up for processing
			req.timer.Stop()
			req.response = pack.(*statePack).states
			req.delivered = time.Now()

			finished = append(finished, req)
			delete(active, pack.PeerId())

		// Handle dropped peer connections:
		case p := <-peerDrop:
			// Skip if no request is currently pending
			req := active[p.id]
			if req == nil {
				continue
			}
			// Finalize the request and queue up for processing
			req.timer.Stop()
			req.dropped = true
			req.delivered = time.Now()

			finished = append(finished, req)
			delete(active, p.id)

		// Handle timed-out requests:
		case req := <-timeout:
			// If the peer is already requesting something else, ignore the stale timeout.
			// This can happen when the timeout and the delivery happens simultaneously,
			// causing both pathways to trigger.
			if active[req.peer.id] != req {
				continue
			}
			req.delivered = time.Now()
			// Move the timed out data back into the download queue
			finished = append(finished, req)
			delete(active, req.peer.id)

		// Track outgoing state requests:
		case req := <-d.trackStateReq:
			// If an active request already exists for this peer, we have a problem. In
			// theory the trie node schedule must never assign two requests to the same
			// peer. In practice however, a peer might receive a request, disconnect and
			// immediately reconnect before the previous times out. In this case the first
			// request is never honored, alas we must not silently overwrite it, as that
			// causes valid requests to go missing and sync to get stuck.
			if old := active[req.peer.id]; old != nil {
				log.Warn("Busy peer assigned new state fetch", "peer", old.peer.id)
				// Move the previous request to the finished set
				old.timer.Stop()
				old.dropped = true
				old.delivered = time.Now()
				finished = append(finished, old)
			}
			// Start a timer to notify the sync loop if the peer stalled.
			req.timer = time.AfterFunc(req.timeout, func() {
				timeout <- req
			})
			active[req.peer.id] = req
		}
	}
}

// spindownStateSync 'drains' the outstanding requests; some will be delivered and other
// will time out. This is to ensure that when the next stateSync starts working, all peers
// are marked as idle and de facto _are_ idle.
func (d *Leecher) spindownStateSync(active map[string]*stateReq, finished []*stateReq, timeout chan *stateReq, peerDrop chan *peerConnection) {
	log.Trace("State sync spinning down", "active", len(active), "finished", len(finished))
	for len(active) > 0 {
		var (
			req    *stateReq
			reason string
		)
		select {
		// Handle (drop) incoming state packs:
		case pack := <-d.stateCh:
			req = active[pack.PeerId()]
			reason = "delivered"
		// Handle dropped peer connections:
		case p := <-peerDrop:
			req = active[p.id]
			reason = "peerdrop"
		// Handle timed-out requests:
		case req = <-timeout:
			reason = "timeout"
		}
		if req == nil {
			continue
		}
		req.peer.log.Trace("State peer marked idle (spindown)", "req.items", int(req.nItems), "reason", reason)
		req.timer.Stop()
		delete(active, req.peer.id)
		req.peer.SetNodeDataIdle(int(req.nItems), time.Now())
	}
	// The 'finished' set contains deliveries that we were going to pass to processing.
	// Those are now moot, but we still need to set those peers as idle, which would
	// otherwise have been done after processing
	for _, req := range finished {
		req.peer.SetNodeDataIdle(int(req.nItems), time.Now())
	}
}

// stateSync schedules requests for downloading a particular state trie defined
// by a given state root.
type stateSync struct {
	d *Leecher // Downloader instance to access and manage current peerset

	root   common.Hash        // State root currently being synced
	sched  *trie.Sync         // State trie sync scheduler defining the tasks
	keccak crypto.KeccakState // Keccak256 hasher to verify deliveries with

	trieTasks map[common.Hash]*trieTask // Set of trie node tasks currently queued for retrieval
	codeTasks map[common.Hash]*codeTask // Set of byte code tasks currently queued for retrieval

	numUncommitted   int
	bytesUncommitted int

	started chan struct{} // Started is signalled once the sync loop starts

	deliver    chan *stateReq // Delivery channel multiplexing peer responses
	cancel     chan struct{}  // Channel to signal a termination request
	cancelOnce sync.Once      // Ensures cancel only ever gets called once
	done       chan struct{}  // Channel to signal termination completion
	err        error          // Any error hit during sync (set before completion)
}

// trieTask represents a single trie node download task, containing a set of
// peers already attempted retrieval from to detect stalled syncs and abort.
type trieTask struct {
	path     [][]byte
	attempts map[string]struct{}
}

// codeTask represents a single byte code download task, containing a set of
// peers already attempted retrieval from to detect stalled syncs and abort.
type codeTask struct {
	attempts map[string]struct{}
}

// newStateSync creates a new state trie download scheduler. This method does not
// yet start the sync. The user needs to call run to initiate.
func newStateSync(d *Leecher, root common.Hash) *stateSync {
	return &stateSync{
		d:         d,
		root:      root,
		sched:     state.NewStateSync(root, d.stateDB, d.stateBloom, nil),
		keccak:    sha3.NewLegacyKeccak256().(crypto.KeccakState),
		trieTasks: make(map[common.Hash]*trieTask),
		codeTasks: make(map[common.Hash]*codeTask),
		deliver:   make(chan *stateReq),
		cancel:    make(chan struct{}),
		done:      make(chan struct{}),
		started:   make(chan struct{}),
	}
}

// run starts the task assignment and response processing loop, blocking until
// it finishes, and finally notifying any goroutines waiting for the loop to
// finish.
func (s *stateSync) run() {
	close(s.started)
	s.err = s.d.SnapSyncer.Sync(s.root, s.cancel)
	close(s.done)
}

// Wait blocks until the sync is done or canceled.
func (s *stateSync) Wait() error {
	<-s.done
	return s.err
}

// Cancel cancels the sync and waits until it has shut down.
func (s *stateSync) Cancel() error {
	s.cancelOnce.Do(func() {
		close(s.cancel)
	})
	return s.Wait()
}
