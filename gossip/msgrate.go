package gossip

import (
	"errors"
	"math"
	"sync"
)

type Tracker struct {
	capacity map[uint64]float64

	lock sync.RWMutex
}

func NewTracker(caps map[uint64]float64) *Tracker {
	if caps == nil {
		caps = make(map[uint64]float64)
	}
	return &Tracker{
		capacity: caps,
	}
}

func (t *Tracker) Update(kind uint64, items int) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.capacity[kind] = float64(items)
}

func (t *Tracker) Capacity(kind uint64) int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return roundCapacity(t.capacity[kind])
}

func roundCapacity(cap float64) int {
	const maxInt32 = float64(1<<31 - 1)
	return int(math.Min(maxInt32, math.Max(1, math.Ceil(cap))))
}

type Trackers struct {
	trackers map[string]*Tracker

	lock sync.RWMutex
}

// NewTrackers creates an empty set of trackers to be filled with peers.
func NewTrackers() *Trackers {
	return &Trackers{
		trackers: make(map[string]*Tracker),
	}
}

// Untrack stops tracking a previously added peer.
func (t *Trackers) Untrack(id string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if _, ok := t.trackers[id]; !ok {
		return errors.New("not tracking")
	}
	delete(t.trackers, id)
	return nil
}

func (t *Trackers) Track(id string, tracker *Tracker) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if _, ok := t.trackers[id]; ok {
		return errors.New("already tracking")
	}
	t.trackers[id] = tracker

	return nil
}

// Update is a helper function to access a specific tracker without having to
// track it explicitly outside.
func (t *Trackers) Update(id string, kind uint64, items int) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if tracker := t.trackers[id]; tracker != nil {
		tracker.Update(kind, items)
	}
}

// MeanCapacities returns the capacities averaged across all the added trackers.
// The purpose of the mean capacities are to initialize a new peer with some sane
// starting values that it will hopefully outperform. If the mean overshoots, the
// peer will be cut back to minimal capacity and given another chance.
func (t *Trackers) MeanCapacities() map[uint64]float64 {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.meanCapacities()
}

// meanCapacities is the internal lockless version of MeanCapacities used for
// debug logging.
func (t *Trackers) meanCapacities() map[uint64]float64 {
	capacities := make(map[uint64]float64)
	for _, tt := range t.trackers {
		tt.lock.RLock()
		for key, val := range tt.capacity {
			capacities[key] += val
		}
		tt.lock.RUnlock()
	}
	for key, val := range capacities {
		capacities[key] = val / float64(len(t.trackers))
	}
	return capacities
}

// Capacity is a helper function to access a specific tracker without having to
// track it explicitly outside.
func (t *Trackers) Capacity(id string, kind uint64) int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	tracker := t.trackers[id]
	if tracker == nil {
		return 1
	}
	return tracker.Capacity(kind)
}

// capacitySort implements the Sort interface, allowing sorting by peer message
// throughput. Note, callers should use sort.Reverse to get the desired effect
// of highest capacity being at the front.
type capacitySort struct {
	ids  []string
	caps []int
}

func (s *capacitySort) Len() int {
	return len(s.ids)
}

func (s *capacitySort) Less(i, j int) bool {
	return s.caps[i] < s.caps[j]
}

func (s *capacitySort) Swap(i, j int) {
	s.ids[i], s.ids[j] = s.ids[j], s.ids[i]
	s.caps[i], s.caps[j] = s.caps[j], s.caps[i]
}
