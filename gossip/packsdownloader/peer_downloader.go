package packsdownloader

import (
	"errors"
	"math"
	"time"

	tree "github.com/emirpasic/gods/maps/treemap"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-lachesis/gossip/fetcher"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

const (
	forceSyncPeriod = 5 * time.Minute        // Even if we synced up, in a case we're stalled, try to download a not pinned pack after the timeout
	arriveTimeout   = 10 * time.Second       // Time allowance before an announced pack is explicitly requested
	recheckInterval = 100 * time.Millisecond // Time between checking - was the arrived pack connected or not

	// Maximum number of stored packs per peer.
	// Shouldn't be high, because we do binary search, so stored packs are O(log_2(total packs)) + PeerProgress broadcasts
	maxPeerPacks = 128
	// Maximum number of parallel full pack requests to a peer
	maxFetchingFullPacks = 3

	// maxQueuedFullPacks is the maximum number of inject batches to queue up before
	// dropping incoming packs.
	maxQueuedFullPacks = 8
	// maxQueuedInfos is the maximum number of announce batches to queue up before
	// dropping incoming pack infos.
	maxQueuedInfos = 32
)

var (
	errTerminated = errors.New("terminated")
)

// onlyNotConnectedFn returns only not connected events.
type onlyNotConnectedFn func(ids hash.Events) hash.Events

// dropPeerFn is a callback type for dropping a peer detected as malicious.
type dropPeerFn func(peer string)

// request pack info from the peer
type packInfoRequesterFn func(epoch idx.Epoch, indexes []idx.Pack) error

// request full pack from the peer
type packRequesterFn func(epoch idx.Epoch, index idx.Pack) error

type packsNumData struct {
	epoch    idx.Epoch // in the specified epoch
	packsNum idx.Pack  // there's this number of packs
}

type packInfoData struct {
	epoch idx.Epoch   // the epoch where pack is located
	index idx.Pack    // the seq number of the pack
	heads hash.Events // Hashes of the pack heads
	time  time.Time   // Timestamp of the announcement
}

type packData struct {
	epoch idx.Epoch   // the epoch where pack is located
	index idx.Pack    // the seq number of the pack
	ids   hash.Events // Event hashes which form the pack
	time  time.Time   // Timestamp of the announcement

	fetchEvents fetcher.EventsRequesterFn
}

// PeerPacksDownloader is responsible for accumulating pack announcements from various peers
// and scheduling them for retrieval.
type PeerPacksDownloader struct {
	// Various event channels
	notifyInfo     chan *packInfoData
	notifyPacksNum chan *packsNumData
	notifyPack     chan *packData

	quit chan struct{}

	// Callbacks
	dropPeer         dropPeerFn
	fetcher          *fetcher.Fetcher
	onlyNotConnected onlyNotConnectedFn

	// Announce states
	myEpoch idx.Epoch // the epoch where where we're syncing
	peer    Peer      // the peer we're syncing with

	packsNum     idx.Pack               // total num of packs the peer has (not all of them are requested!)
	packInfos    *tree.Map              // the short descriptors of received peer's packs
	fetchingInfo map[idx.Pack]time.Time // the packs we've requested
	fetchingFull map[idx.Pack]time.Time // the packs we've requested
	prevRequest  time.Time              // time of prev. request to the peer
}

// New creates a packs fetcher to retrieve events based on pack announcements. Works only with 1 peer.
func newPeer(peer Peer, myEpoch idx.Epoch, fetcher *fetcher.Fetcher, onlyNotConnected onlyNotConnectedFn, dropPeer dropPeerFn) *PeerPacksDownloader {
	return &PeerPacksDownloader{
		notifyInfo:       make(chan *packInfoData, maxQueuedInfos),
		notifyPacksNum:   make(chan *packsNumData, maxQueuedInfos),
		notifyPack:       make(chan *packData, maxQueuedFullPacks),
		quit:             make(chan struct{}),
		packInfos:        tree.NewWithIntComparator(),
		fetchingInfo:     make(map[idx.Pack]time.Time),
		fetchingFull:     make(map[idx.Pack]time.Time),
		peer:             peer,
		myEpoch:          myEpoch,
		fetcher:          fetcher,
		onlyNotConnected: onlyNotConnected,
		dropPeer:         dropPeer,
	}
}

// Start boots up the announcement based synchroniser, accepting and processing
// hash notifications and event fetches until termination requested.
func (d *PeerPacksDownloader) Start() {
	go d.loop()
}

// Stop terminates the announcement based synchroniser, canceling all pending
// operations.
func (d *PeerPacksDownloader) Stop() {
	close(d.quit)
}

// NotifyPackInfo injects new pack infos from a peer
func (d *PeerPacksDownloader) NotifyPackInfo(epoch idx.Epoch, index idx.Pack, heads hash.Events, time time.Time) error {
	if d.myEpoch != epoch {
		return nil // Short circuit if from another epoch
	}

	op := &packInfoData{
		epoch: epoch,
		index: index,
		heads: heads,
		time:  time,
	}
	select {
	case d.notifyInfo <- op:
		return nil
	case <-d.quit:
		return errTerminated
	}
}

// NotifyPacksNum injects new packs num from a peer
func (d *PeerPacksDownloader) NotifyPacksNum(epoch idx.Epoch, packsNum idx.Pack) error {
	if d.myEpoch != epoch {
		return nil // Short circuit if from another epoch
	}

	op := &packsNumData{
		epoch:    epoch,
		packsNum: packsNum,
	}
	select {
	case d.notifyPacksNum <- op:
		return nil
	case <-d.quit:
		return errTerminated
	}
}

// NotifyPack injects new packs from a peer
func (d *PeerPacksDownloader) NotifyPack(epoch idx.Epoch, index idx.Pack, ids hash.Events, time time.Time, fetchEvents fetcher.EventsRequesterFn) error {
	if d.myEpoch != epoch {
		return nil // Short circuit if from another epoch
	}

	op := &packData{
		epoch:       epoch,
		index:       index,
		ids:         ids,
		time:        time,
		fetchEvents: fetchEvents,
	}
	select {
	case d.notifyPack <- op:
		return nil
	case <-d.quit:
		return errTerminated
	}
}

// Loop is the main downloader's loop, checking and processing various notifications
func (d *PeerPacksDownloader) loop() {
	// Iterate the event fetching until a quit is requested
	syncTicker := time.NewTicker(recheckInterval)

	for {
		// Wait for an outside event to occur
		select {
		case <-d.quit:
			// PeerPacksDownloader terminating, abort all operations
			return

		case op := <-d.notifyPacksNum:
			if d.myEpoch != op.epoch {
				continue // from another epoch
			}
			if d.packsNum < op.packsNum {
				d.packsNum = op.packsNum
			}

		case packInfo := <-d.notifyInfo:
			if d.myEpoch != packInfo.epoch {
				continue // from another epoch
			}
			if d.packsNum < packInfo.index {
				d.packsNum = packInfo.index
			}

			if d.packInfos.Size() > maxPeerPacks {
				// if we have too much packs -> d.sweepKnown() doesn't erase them -> we don't connect events from these packs.
				// Also we do binary search, so we don't need much packs to store, so we shouldn't reach this if peer is ok.
				log.Warn("All the peer packs are unknown. Faulty peer?", "peer", d.peer.ID)
				d.dropPeer(d.peer.ID)
			}
			if packInfo.index <= 0 || packInfo.index >= math.MaxInt32 {
				log.Error("Invalid pack index", "peer", d.peer.ID)
				continue
			}

			d.packInfos.Put(int(packInfo.index), packInfo)
			delete(d.fetchingInfo, packInfo.index) // mark as received

			d.tryToSync()
			d.sweepKnown()

		case pack := <-d.notifyPack:
			if d.myEpoch != pack.epoch {
				continue // from another epoch
			}
			if d.packsNum < pack.index {
				d.packsNum = pack.index
			}

			// DO NOT erase from d.fetchingFull!
			// We should erase only when we'll actually connect all the events form this pack, i.e. when pack's heads are known.
			// It'll be done in sweepKnown()
			// Otherwise, we'll rapidly re-request the same pack until we connect events.
			// DO NOT: delete(d.fetchingFull, pack.index)

			err := d.fetcher.Notify(d.peer.ID, pack.ids, pack.time, pack.fetchEvents)
			if err != nil {
				log.Error("Pack inject error", "index", pack.index, "peer", d.peer.ID, "err", err)
			}

		case <-syncTicker.C:
			d.tryToSync()
			d.sweepKnown()
		}
	}
}

func (d *PeerPacksDownloader) tryToSync() {
	if d.fetcher.OverloadedPeer(d.peer.ID) {
		return
	}

	index, requestFull, syncedUp := d.binarySearchReq()
	if syncedUp {
		// even if we're synced up, in a case we're stalled, try to download a not pinned pack after the timeout
		stalled := time.Since(d.prevRequest) > forceSyncPeriod
		if stalled {
			d.timedRequestPackInfo(d.packsNum + 1)
			d.timedRequestFullPack(d.packsNum+1, false)
		}
		return
	}

	if requestFull {
		// request a few packs in parallel
		for i := index; i < index+maxFetchingFullPacks && i <= d.packsNum; i++ {
			d.timedRequestFullPack(i, true)
			if i+1 <= d.packsNum {
				_, found := d.packInfos.Get(int(i + 1))
				if !found {
					d.timedRequestPackInfo(i + 1) // request new pack info in advance
				}
			}
		}
	} else {
		d.timedRequestPackInfo(index)
	}
}

// Wrapper does the request only if passed enough time since prev request
// If pack isn't pinned, then it'll be different every time we request, so we must not remember it
func (d *PeerPacksDownloader) timedRequestFullPack(index idx.Pack, pinned bool) {
	prevRequestTime := d.fetchingFull[index]
	if prevRequestTime.IsZero() || time.Since(prevRequestTime) >= arriveTimeout {
		err := d.peer.RequestPack(d.myEpoch, index)
		if err != nil {
			log.Warn("Pack request error", "index", index, "peer", d.peer.ID, "err", err)
		}
		d.prevRequest = time.Now()
		if pinned {
			d.fetchingFull[index] = d.prevRequest
		}
	}
}

// Wrapper does the request only if passed enough time since prev request
func (d *PeerPacksDownloader) timedRequestPackInfo(index idx.Pack) {
	prevRequestTime := d.fetchingInfo[index]
	if prevRequestTime.IsZero() || time.Since(prevRequestTime) >= arriveTimeout {
		err := d.peer.RequestPackInfos(d.myEpoch, []idx.Pack{index})
		if err != nil {
			log.Warn("Pack info request error", "index", index, "peer", d.peer.ID, "err", err)
		}
		d.prevRequest = time.Now()
		d.fetchingInfo[index] = d.prevRequest
	}
}

// Finds lowest not known pack, such that previous pack is known or doesn't exist
// If not found, returns next pack index to request
func (d *PeerPacksDownloader) binarySearchReq() (requestIndex idx.Pack, requestFull bool, syncedUp bool) {
	it := d.packInfos.Iterator()
	var prevIdx *idx.Pack

	for it.End(); it.Prev(); {
		packInfo := it.Value().(*packInfoData)

		allKnown := len(d.onlyNotConnected(packInfo.heads)) == 0
		packIdx := idx.Pack(it.Key().(int))

		if allKnown {
			var nextReq idx.Pack
			if prevIdx == nil {
				// ..., known 5
				// the first pack is known
				if packIdx >= d.packsNum {
					// ..., known 5.
					// we're already synced up
					return packIdx, false, true
				}
				nextReq = packIdx + (d.packsNum-packIdx)/2
			} else {
				if packIdx+1 == *prevIdx {
					// ..., known 5, not known 6, not known 12, not known 15, ...
					// fetch the lowest not known pack
					return packIdx + 1, true, false
				}
				nextReq = packIdx + (*prevIdx-packIdx)/2
			}
			if nextReq <= packIdx {
				nextReq = packIdx + 1
			}
			// ..., known 5, not known 8, not known 12, not known 15, ...
			// binary search to find lowest not known pack next time
			return nextReq, false, false
		}
		prevIdx = &packIdx
	}
	// if we're here, then no known packs were found
	if prevIdx == nil {
		return 1, false, false
	}
	if *prevIdx == 1 {
		// if 1 pack isn't known, then it's the lowest not known
		return 1, true, false
	}
	return *prevIdx / 2, false, false
}

// Erases all the pack infos before highest known pack info (because they will be known too)
func (d *PeerPacksDownloader) sweepKnown() {
	it := d.packInfos.Iterator()
	toRemove := make([]idx.Pack, 0, d.packInfos.Size())
	allKnownMet := false

	for it.End(); it.Prev(); {
		packIdx := idx.Pack(it.Key().(int))
		packInfo := it.Value().(*packInfoData)
		allKnown := len(d.onlyNotConnected(packInfo.heads)) == 0
		if allKnownMet {
			toRemove = append(toRemove, packIdx)

			if !allKnown {
				log.Error("Peer downloader error: met pack with an unknown head before pack with only known heads. Faulty peer?", "peer", d.peer.ID)
			}
		} else if allKnown {
			allKnownMet = true
		}
	}

	for _, r := range toRemove {
		d.forgetPack(r)
	}
}

// forgetPack removes all traces of a pack announcement from the fetcher's
// internal state.
func (d *PeerPacksDownloader) forgetPack(index idx.Pack) {
	d.packInfos.Remove(int(index))
	delete(d.fetchingInfo, index)
	delete(d.fetchingFull, index)
}
