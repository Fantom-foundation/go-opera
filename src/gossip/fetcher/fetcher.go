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
	fetchTimeout  = 5 * time.Second        // Maximum allotted time to return an explicitly requested event
	hashLimit     = 2048                   // Maximum number of unique events a peer may have announced
)

var (
	errTerminated = errors.New("terminated")
)

// dropPeerFn is a callback type for dropping a peer detected as malicious.
type dropPeerFn func(peer string)

// eventsRequesterFn is a callback type for sending a event retrieval request.
type eventsRequesterFn func([]hash.Event) error

// inject represents a schedules import operation.
type inject struct {
	peer   string
	events []*inter.Event
}

// announces is the hash notification of the availability of new events in the
// network.
type announcesBatch struct {
	hashes []hash.Event // Hashes of the events being announced
	time   time.Time    // Timestamp of the announcement

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

	pushEvent    ordering.PushEventFn
	isEventKnown ordering.IsBufferedFn
	dropPeer     dropPeerFn

	quit chan struct{}

	// Announce states
	announces map[string]int                // Per peer announce counts to prevent memory exhaustion
	announced map[hash.Event][]*oneAnnounce // Announced events, scheduled for fetching
	fetching  map[hash.Event]*oneAnnounce   // Announced events, currently fetching
}

// New creates a event fetcher to retrieve events based on hash announcements.
func New(pushEvent ordering.PushEventFn, isEventKnown ordering.IsBufferedFn, dropPeer dropPeerFn) *Fetcher {
	return &Fetcher{
		notify:       make(chan *announcesBatch),
		inject:       make(chan *inject),
		quit:         make(chan struct{}),
		announces:    make(map[string]int),
		announced:    make(map[hash.Event][]*oneAnnounce),
		fetching:     make(map[hash.Event]*oneAnnounce),
		pushEvent:    pushEvent,
		isEventKnown: isEventKnown,
		dropPeer:     dropPeer,
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
func (f *Fetcher) Notify(peer string, hashes []hash.Event, time time.Time, fetchEvents eventsRequesterFn) error {
	events := &announcesBatch{
		hashes:      hashes,
		time:        time,
		peer:        peer,
		fetchEvents: fetchEvents,
	}
	select {
	case f.notify <- events:
		return nil
	case <-f.quit:
		return errTerminated
	}
}

// Enqueue tries to fill gaps the fetcher's future import queue.
func (f *Fetcher) Enqueue(peer string, events []*inter.Event) error {
	op := &inject{
		peer:   peer,
		events: events,
	}
	select {
	case f.inject <- op:
		return nil
	case <-f.quit:
		return errTerminated
	}
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

			count := f.announces[notification.peer] + len(notification.hashes)
			if count > hashLimit {
				log.Debug("Peer exceeded outstanding announces", "peer", notification.peer, "limit", hashLimit)
				propAnnounceDOSMeter.Update(1)
				break
			}

			// Schedule all the unknown hashes for retrieval
			unknown := make([]hash.Event, 0, len(notification.hashes))
			for i, id := range notification.hashes {
				if _, ok := f.fetching[id]; ok {
					continue
				}
				if !f.isEventKnown(id) {
					unknown = append(unknown, id)
				}
				f.announced[id] = append(f.announced[id], &oneAnnounce{
					batch: notification,
					i:     i,
				})
			}

			err := notification.fetchEvents(unknown)
			if err != nil {
				log.Error("Events request error", "peer", notification.peer, "err", err)
			}

			f.announces[notification.peer] = count
			if len(f.announced) == 1 {
				f.rescheduleFetch(fetchTimer)
			}

		case op := <-f.inject:
			// A direct event insertion was requested, try and fill any pending gaps
			propBroadcastInMeter.Update(1)
			for _, e := range op.events {
				f.pushEvent(e, op.peer)
				f.forgetHash(e.Hash())
			}

		case <-fetchTimer.C:
			// At least one event's timer ran out, check for needing retrieval
			request := make(map[string][]hash.Event)

			for hash, announces := range f.announced {
				if time.Since(announces[0].batch.time) > arriveTimeout-gatherSlack {
					// Pick a random peer to retrieve from, reset all others
					announce := announces[rand.Intn(len(announces))]
					f.forgetHash(hash)

					// If the event still didn't arrive, queue for fetching
					if !f.isEventKnown(hash) {
						request[announce.batch.peer] = append(request[announce.batch.peer], hash)
						f.fetching[hash] = announce
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
