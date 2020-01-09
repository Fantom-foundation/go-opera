package fetcher

import (
	"errors"
	"math/rand"
	"time"

	"github.com/Fantom-foundation/go-lachesis/eventcheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/heavycheck"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/logger"
	"github.com/Fantom-foundation/go-lachesis/utils"
)

/*
 * Fetcher is a network agent, which handles basic hash-based events sync.
 * The core mechanic is very simple: interested hash arrived => request it.
 * The main reason why it has more than a few lines of code,
 * is because it tries to protect itself (and other nodes) against DoS.
 */

const (
	forgetTimeout = 1 * time.Minute         // Time before an announced event is forgotten
	arriveTimeout = 1000 * time.Millisecond // Time allowance before an announced event is explicitly requested
	gatherSlack   = 100 * time.Millisecond  // Interval used to collate almost-expired announces with fetches
	fetchTimeout  = 10 * time.Second        // Maximum allowed time to return an explicitly requested event
	hashLimit     = 3000                    // Maximum number of unique events a peer may have announced

	maxInjectBatch   = 4   // Maximum number of events in an inject batch (batch is divided if exceeded)
	maxAnnounceBatch = 256 // Maximum number of hashes in an announce batch (batch is divided if exceeded)

	// maxQueuedInjects is the maximum number of inject batches to queue up before
	// dropping incoming events.
	maxQueuedInjects = 128
	// maxQueuedAnns is the maximum number of announce batches to queue up before
	// dropping incoming hashes.
	maxQueuedAnns = 128
)

var (
	errTerminated = errors.New("terminated")
)

// DropPeerFn is a callback type for dropping a peer detected as malicious.
type DropPeerFn func(peer string)

// FilterInterestedFn returns only event which may be requested.
type FilterInterestedFn func(ids hash.Events) hash.Events

// EventsRequesterFn is a callback type for sending a event retrieval request.
type EventsRequesterFn func(hash.Events) error

// PushEventFn is a callback type to connect a received event
type PushEventFn func(e *inter.Event, peer string)

// inject represents a schedules import operation.
type inject struct {
	events []*inter.Event // Incoming events
	time   time.Time      // Timestamp when received

	peer string // Identifier of the peer which sent events

	fetchEvents EventsRequesterFn
}

// announces is the hash notification of the availability of new events in the
// network.
type announcesBatch struct {
	hashes hash.Events // Hashes of the events being announced
	time   time.Time   // Timestamp of the announcement

	peer string // Identifier of the peer originating the notification

	fetchEvents EventsRequesterFn
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
	quit   chan struct{}

	// Callbacks
	callback Callback

	// Announce states
	stateMu   utils.SpinLock                // Protects announces and announced
	announces map[string]int                // Per peer announce counts to prevent memory exhaustion
	announced map[hash.Event][]*oneAnnounce // Announced events, scheduled for fetching

	fetching     map[hash.Event]*oneAnnounce // Announced events, currently fetching
	fetchingTime map[hash.Event]time.Time

	logger.Periodic
}

type Callback struct {
	PushEvent      PushEventFn
	OnlyInterested FilterInterestedFn
	DropPeer       DropPeerFn

	HeavyCheck *heavycheck.Checker
	FirstCheck func(*inter.Event) error
}

// New creates a event fetcher to retrieve events based on hash announcements.
func New(callback Callback) *Fetcher {
	loggerInstance := logger.MakeInstance()
	return &Fetcher{
		notify:       make(chan *announcesBatch, maxQueuedAnns),
		inject:       make(chan *inject, maxQueuedInjects),
		quit:         make(chan struct{}),
		announces:    make(map[string]int),
		announced:    make(map[hash.Event][]*oneAnnounce),
		fetching:     make(map[hash.Event]*oneAnnounce),
		fetchingTime: make(map[hash.Event]time.Time),
		callback:     callback,

		Periodic: logger.Periodic{Instance: loggerInstance},
	}
}

// Start boots up the announcement based synchroniser, accepting and processing
// hash notifications and event fetches until termination requested.
func (f *Fetcher) Start() {
	f.callback.HeavyCheck.Start()
	go f.loop()
}

// Stop terminates the announcement based synchroniser, canceling all pending
// operations.
func (f *Fetcher) Stop() {
	close(f.quit)
	f.callback.HeavyCheck.Stop()
}

// Overloaded returns true if too much events are being processed or requested
func (f *Fetcher) Overloaded() bool {
	f.stateMu.Lock()
	defer f.stateMu.Unlock()
	return f.overloaded()
}

func (f *Fetcher) overloaded() bool {
	return len(f.inject) > maxQueuedInjects*3/4 ||
		len(f.notify) > maxQueuedAnns*3/4 ||
		len(f.announced) > hashLimit || // protected by stateMu
		f.callback.HeavyCheck.Overloaded()
}

// OverloadedPeer returns true if too much events are being processed or requested from the peer
func (f *Fetcher) OverloadedPeer(peer string) bool {
	f.stateMu.Lock()
	defer f.stateMu.Unlock()
	return f.overloaded() || f.announces[peer] > hashLimit/2 // protected by stateMu
}

func (f *Fetcher) setAnnounces(peer string, num int) {
	f.stateMu.Lock()
	defer f.stateMu.Unlock()
	f.announces[peer] = num
}

func (f *Fetcher) setAnnounced(id hash.Event, announces []*oneAnnounce) {
	f.stateMu.Lock()
	defer f.stateMu.Unlock()
	f.announced[id] = announces
}

// Notify announces the fetcher of the potential availability of a new event in
// the network.
func (f *Fetcher) Notify(peer string, hashes hash.Events, time time.Time, fetchEvents EventsRequesterFn) error {
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
func (f *Fetcher) Enqueue(peer string, inEvents inter.Events, t time.Time, fetchEvents EventsRequesterFn) error {
	// Filter already known events
	notKnownEvents := make(inter.Events, 0, len(inEvents))
	for _, e := range inEvents {
		if len(f.callback.OnlyInterested(hash.Events{e.Hash()})) == 0 {
			continue
		}
		notKnownEvents = append(notKnownEvents, e)
	}

	// Run light checks right away
	passed := make(inter.Events, 0, len(notKnownEvents))
	for _, e := range notKnownEvents {
		err := f.callback.FirstCheck(e)
		if eventcheck.IsBan(err) {
			f.Periodic.Warn(time.Second, "Incoming event rejected", "event", e.Hash().String(), "creator", e.Creator, "err", err)
			f.callback.DropPeer(peer)
			return err
		}
		if err == nil {
			passed = append(passed, e)
		}
	}

	// Run heavy check in parallel
	return f.callback.HeavyCheck.Enqueue(passed, func(res *heavycheck.TaskData) {
		// Check errors of heavy check
		passed := make(inter.Events, 0, len(res.Events))
		for i, err := range res.Result {
			if eventcheck.IsBan(err) {
				e := res.Events[i]
				f.Periodic.Warn(time.Second, "Incoming event rejected", "event", e.Hash().String(), "creator", e.Creator, "err", err)
				f.callback.DropPeer(peer)
				return
			}
			if err == nil {
				passed = append(passed, res.Events[i])
			}
		}
		// after all these checks, actually enqueue the events into fetcher
		_ = f.enqueue(peer, passed, t, fetchEvents)
	})
}

func (f *Fetcher) enqueue(peer string, events inter.Events, time time.Time, fetchEvents EventsRequesterFn) error {
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

// Loop is the main fetcher loop, checking and processing various notifications
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
			propAnnounceInMeter.Update(int64(len(notification.hashes)))

			count := f.announces[notification.peer]
			if count+len(notification.hashes) > hashLimit {
				f.Periodic.Debug(time.Second, "Peer exceeded outstanding announces", "peer", notification.peer, "limit", hashLimit)
				propAnnounceDOSMeter.Update(1)
				break
			}

			first := len(f.fetching) == 0

			// filter only not known
			notification.hashes = f.callback.OnlyInterested(notification.hashes)
			if len(notification.hashes) == 0 {
				break
			}

			toFetch := make(hash.Events, 0, len(notification.hashes))
			for i, id := range notification.hashes {
				// add new announcement. other peers may already have announced it, so it's an array
				ann := &oneAnnounce{
					batch: notification,
					i:     i,
				}
				f.setAnnounced(id, append(f.announced[id], ann))
				count++ // f.announced and f.announces must be synced!
				// if it wasn't announced before, then schedule for fetching this time
				if _, ok := f.fetching[id]; !ok {
					f.fetching[id] = ann
					f.fetchingTime[id] = notification.time
					toFetch.Add(id)
				}
			}
			f.setAnnounces(notification.peer, count)

			if len(toFetch) != 0 {
				err := notification.fetchEvents(toFetch)
				if err != nil {
					f.Periodic.Warn(time.Second, "Events request error", "peer", notification.peer, "err", err)
				}
			}

			if first && len(f.fetching) != 0 {
				f.rescheduleFetch(fetchTimer)
			}

		case op := <-f.inject:
			// A direct event insertion was requested, try and fill any pending gaps
			parents := make(hash.Events, 0, len(op.events))
			propBroadcastInMeter.Update(int64(len(op.events)))
			for _, e := range op.events {
				// fetch unknown parents
				for _, p := range e.Parents {
					if _, ok := f.fetching[p]; ok {
						continue
					}
					parents.Add(p)
				}

				f.callback.PushEvent(e, op.peer)
				f.forgetHash(e.Hash())
			}

			parents = f.callback.OnlyInterested(parents)
			if len(parents) != 0 && !f.OverloadedPeer(op.peer) {
				// f.Notify will filter onlyInterested parents - this way, we won't request the events from op.events
				_ = f.Notify(op.peer, parents, op.time, op.fetchEvents)
			}

		case now := <-fetchTimer.C:
			// At least one event's timer ran out, check for needing retrieval
			request := make(map[string]hash.Events)

			// Find not not arrived events
			all := make(hash.Events, 0, len(f.announced))
			for e := range f.announced {
				all.Add(e)
			}
			notArrived := f.callback.OnlyInterested(all)

			for _, e := range notArrived {
				// Re-fetch not arrived events
				announces := f.announced[e]

				oldest := announces[0] // first is the oldest
				if time.Since(oldest.batch.time) > forgetTimeout {
					// Forget too old announces
					f.forgetHash(e)
				} else if time.Since(f.fetchingTime[e]) > arriveTimeout-gatherSlack {
					// The event still didn't arrive, queue for fetching from a random peer
					announce := announces[rand.Intn(len(announces))]
					request[announce.batch.peer] = append(request[announce.batch.peer], e)
					f.fetching[e] = announce
					f.fetchingTime[e] = now
				}
			}

			// Forget arrived events.
			// It's possible to get here only if event arrived out-of-fetcher, via another channel.
			// Also may be possible after a change of an epoch.
			notArrivedM := notArrived.Set()
			for _, e := range all {
				if !notArrivedM.Contains(e) {
					f.forgetHash(e)
				}
			}

			// Send out all event requests
			for peer, hashes := range request {
				f.Log.Trace("Fetching scheduled events", "peer", peer, "count", len(hashes))

				// Create a closure of the fetch and schedule in on a new thread
				fetchEvents, hashes := f.fetching[hashes[0]].batch.fetchEvents, hashes
				go func(peer string) {
					eventFetchMeter.Update(int64(len(hashes)))
					err := fetchEvents(hashes)
					if err != nil {
						f.Periodic.Warn(time.Second, "Events request error", "peer", peer, "err", err)
					}
				}(peer)
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
	for _, t := range f.fetchingTime {
		if earliest.After(t) {
			earliest = t
		}
	}
	fetch.Reset(arriveTimeout - time.Since(earliest))
}

// forgetHash removes all traces of a event announcement from the fetcher's
// internal state.
func (f *Fetcher) forgetHash(hash hash.Event) {
	f.stateMu.Lock()
	defer f.stateMu.Unlock()

	// Remove all pending announces and decrement DOS counters
	for _, announce := range f.announced[hash] {
		f.announces[announce.batch.peer]--
		if f.announces[announce.batch.peer] <= 0 {
			delete(f.announces, announce.batch.peer)
		}
	}
	delete(f.announced, hash)
	delete(f.fetching, hash)
	delete(f.fetchingTime, hash)
}
