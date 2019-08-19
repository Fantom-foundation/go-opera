package fetcher

import (
	"errors"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-lachesis/src/gossip/ordering"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

const (
	arriveTimeout = 500 * time.Millisecond // Time allowance before an announced event is explicitly requested
	gatherSlack   = 100 * time.Millisecond // Interval used to collate almost-expired announces with fetches
	fetchTimeout  = 10 * time.Second       // Maximum allotted time to return an explicitly requested event
	hashLimit     = 4096                   // Maximum number of unique events a peer may have announced

	maxInjectBatch   = 8  // Maximum number of events in an inject batch (batch is divided if exceeded)
	maxAnnounceBatch = 16 // Maximum number of hashes in an announce batch (batch is divided if exceeded)

	// maxQueuedInjects is the maximum number of inject batches to queue up before
	// dropping incoming events.
	maxQueuedInjects = 64
	// maxQueuedAnns is the maximum number of announce batches to queue up before
	// dropping incoming hashes.
	maxQueuedAnns = 64
)

var (
	errTerminated = errors.New("terminated")
)

// dropPeerFn is a callback type for dropping a peer detected as malicious.
type dropPeerFn func(peer string)

// isInteresedFn returns true if event may be requested.
type filterInterestedFn func(ids hash.Events) hash.Events

// eventsRequesterFn is a callback type for sending a event retrieval request.
type eventsRequesterFn func(hash.Events) error

// inject represents a schedules import operation.
type inject struct {
	events []*inter.Event // Incoming events
	time   time.Time      // Timestamp when received

	peer string // Identifier of the peer which sent events

	fetchEvents eventsRequesterFn
}

// announces is the hash notification of the availability of new events in the
// network.
type announcesBatch struct {
	hashes hash.Events // Hashes of the events being announced
	time   time.Time   // Timestamp of the announcement

	peer string // Identifier of the peer originating the notification

	fetchEvents eventsRequesterFn
}
type oneAnnounce struct {
	batch *announcesBatch
	i     int
}

// Fetcher is responsible for accumulating event announcements from various peers
// and scheduling them for retrieval.
type Fetcher struct {
	// Various event channels
	notify chan *announcesBatch
	inject chan *inject

	pushEvent        ordering.PushEventFn
	filterInterested filterInterestedFn
	dropPeer         dropPeerFn

	quit chan struct{}

	// Announce states
	announces map[string]int                // Per peer announce counts to prevent memory exhaustion
	announced map[hash.Event][]*oneAnnounce // Announced events, scheduled for fetching
	fetching  map[hash.Event]*oneAnnounce   // Announced events, currently fetching
}

// New creates a event fetcher to retrieve events based on hash announcements.
func New(pushEvent ordering.PushEventFn, filterInterested filterInterestedFn, dropPeer dropPeerFn) *Fetcher {
	return &Fetcher{
		notify:           make(chan *announcesBatch, maxQueuedAnns),
		inject:           make(chan *inject, maxQueuedInjects),
		quit:             make(chan struct{}),
		announces:        make(map[string]int),
		announced:        make(map[hash.Event][]*oneAnnounce),
		fetching:         make(map[hash.Event]*oneAnnounce),
		pushEvent:        pushEvent,
		filterInterested: filterInterested,
		dropPeer:         dropPeer,
	}
}

// Start boots up the announcement based synchroniser, accepting and processing
// hash notifications and event fetches until termination requested.
func (f *Fetcher) Start() {
	go f.loop()
}

// Stop terminates the announcement based synchroniser, canceling all pending
// operations.
func (f *Fetcher) Stop() {
	close(f.quit)
}

// Notify announces the fetcher of the potential availability of a new event in
// the network.
func (f *Fetcher) Notify(peer string, hashes hash.Events, time time.Time, fetchEvents eventsRequesterFn) error {
	// Filter this goroutine to unload the fetcher
	hashes = f.filterInterested(hashes)

	// divide big batch into smaller ones
	for start := 0; start < len(hashes); start += maxAnnounceBatch {
		end := len(hashes)
		if end > start+maxAnnounceBatch {
			end = start + maxAnnounceBatch
		}
		op := &announcesBatch{
			hashes:      hashes[start:end],
			time:        time,
			peer:        peer,
			fetchEvents: fetchEvents,
		}
		select {
		case f.notify <- op:
			continue
		case <-f.quit:
			return errTerminated
		}
	}
	return nil
}

// Enqueue tries to fill gaps the fetcher's future import queue.
func (f *Fetcher) Enqueue(peer string, events []*inter.Event, time time.Time, fetchEvents eventsRequesterFn) error {
	// divide big batch into smaller ones
	for start := 0; start < len(events); start += maxInjectBatch {
		end := len(events)
		if end > start+maxInjectBatch {
			end = start + maxInjectBatch
		}
		op := &inject{
			events:      events[start:end],
			time:        time,
			peer:        peer,
			fetchEvents: fetchEvents,
		}
		select {
		case f.inject <- op:
			continue
		case <-f.quit:
			return errTerminated
		}
	}
	return nil
}

// Loop is the main fetcher loop, checking and processing various notification
// events.
func (f *Fetcher) loop() {
	// Iterate the event fetching until a quit is requested
	fetchTimer := time.NewTimer(0)

	for {
		// Clean up any expired event fetches
		for id, announce := range f.fetching {
			if time.Since(announce.batch.time) > fetchTimeout {
				f.forgetHash(id)
			}
		}
		// Wait for an outside event to occur
		select {
		case <-f.quit:
			// Fetcher terminating, abort all operations
			return

		case notification := <-f.notify:
			// A event was announced, make sure the peer isn't DOSing us
			propAnnounceInMeter.Update(1)

			count := f.announces[notification.peer]
			if count+len(notification.hashes) > hashLimit {
				log.Debug("Peer exceeded outstanding announces", "peer", notification.peer, "limit", hashLimit)
				propAnnounceDOSMeter.Update(1)
				break
			}

			first := len(f.announced) == 0

			// Exclude already fetching
			interested := make(hash.Events, 0, len(notification.hashes))
			for _, id := range notification.hashes {
				if _, ok := f.fetching[id]; ok {
					continue
				}
				interested.Add(id)
			}
			// assume hashes are already filtered
			// interested = f.filterInterested(interested)

			// Fetch interested hashes
			for i, id := range interested {
				f.announced[id] = append(f.announced[id], &oneAnnounce{
					batch: notification,
					i:     i,
				})
				count++ // f.announced and f.announces must be synced!
			}
			f.announces[notification.peer] = count

			if len(interested) != 0 {
				err := notification.fetchEvents(interested)
				if err != nil {
					log.Error("Events request error", "peer", notification.peer, "err", err)
				}
			}

			if first && len(f.announced) != 0 {
				f.rescheduleFetch(fetchTimer)
			}

		case op := <-f.inject:
			// A direct event insertion was requested, try and fill any pending gaps
			unknownParents := make(hash.Events, 0, len(op.events))
			propBroadcastInMeter.Update(1)
			for _, e := range op.events {
				// fetch unknown parents
				for _, p := range e.Parents {
					if _, ok := f.fetching[p]; ok {
						continue
					}
					unknownParents.Add(p)
				}

				f.pushEvent(e, op.peer)
				f.forgetHash(e.Hash())
			}
			// filter after we pushed - this way, we won't request the events from op.events
			unknownParents = f.filterInterested(unknownParents)

			if len(unknownParents) != 0 {
				log.Trace("Fetching events by parents", "peer", op.peer, "count", len(unknownParents))
				_ = f.Notify(op.peer, unknownParents, op.time, op.fetchEvents)
			}

		case <-fetchTimer.C:
			// At least one event's timer ran out, check for needing retrieval
			request := make(map[string]hash.Events)

			for e, announces := range f.announced {
				if time.Since(announces[0].batch.time) > arriveTimeout-gatherSlack {
					// Pick a random peer to retrieve from, reset all others
					announce := announces[rand.Intn(len(announces))]
					f.forgetHash(e)

					// If the event still didn't arrive, queue for fetching
					isInterested := len(f.filterInterested(hash.Events{e})) != 0
					if isInterested {
						request[announce.batch.peer] = append(request[announce.batch.peer], e)
						f.fetching[e] = announce
					}
				}
			}
			// Send out all event requests
			for peer, hashes := range request {
				log.Trace("Fetching scheduled events", "peer", peer, "list", hashes)

				// Create a closure of the fetch and schedule in on a new thread
				fetchEvents, hashes := f.fetching[hashes[0]].batch.fetchEvents, hashes
				go func() {
					eventFetchMeter.Update(int64(len(hashes)))
					err := fetchEvents(hashes)
					if err != nil {
						log.Error("Events request error", "peer", peer, "err", err)
					}
				}()
			}
			// Schedule the next fetch if events are still pending
			f.rescheduleFetch(fetchTimer)
		}
	}
}

// rescheduleFetch resets the specified fetch timer to the next announce timeout.
func (f *Fetcher) rescheduleFetch(fetch *time.Timer) {
	// Short circuit if no events are announced
	if len(f.announced) == 0 {
		return
	}
	// Otherwise find the earliest expiring announcement
	earliest := time.Now()
	for _, announces := range f.announced {
		if earliest.After(announces[0].batch.time) {
			earliest = announces[0].batch.time
		}
	}
	fetch.Reset(arriveTimeout - time.Since(earliest))
}

// forgetHash removes all traces of a event announcement from the fetcher's
// internal state.
func (f *Fetcher) forgetHash(hash hash.Event) {
	// Remove all pending announces and decrement DOS counters
	for _, announce := range f.announced[hash] {
		f.announces[announce.batch.peer]--
		if f.announces[announce.batch.peer] <= 0 {
			delete(f.announces, announce.batch.peer)
		}
	}
	delete(f.announced, hash)
	// Remove any pending fetches and decrement the DOS counters
	if announce := f.fetching[hash]; announce != nil {
		f.announces[announce.batch.peer]--
		if f.announces[announce.batch.peer] <= 0 {
			delete(f.announces, announce.batch.peer)
		}
		delete(f.fetching, hash)
	}
}
