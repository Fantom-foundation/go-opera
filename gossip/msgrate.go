package gossip

import (
	"errors"
	"math"
	"sync"
)

type Tracker struct {
	broadcastedEventRate float64

	lock sync.RWMutex
}

func (t *Tracker) Update(items float64) {
	t.broadcastedEventRate = items
}

func NewTracker(items float64) *Tracker {
	return &Tracker{
		broadcastedEventRate: items,
	}
}

func (t *Tracker) Capacity() int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return roundCapacity(t.broadcastedEventRate)
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

func (t *Trackers) Update(id string, items float64) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if tracker := t.trackers[id]; tracker != nil {
		tracker.Update(items)
	}
}

func (t *Trackers) MeanCapacity() float64 {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.meanCapacity()
}

func (t *Trackers) meanCapacity() float64 {
	capacity := 0.0
	for _, tt := range t.trackers {
		tt.lock.RLock()
		capacity += tt.broadcastedEventRate
		tt.lock.RUnlock()
	}
	return capacity / float64(len(t.trackers))
}

func (t *Trackers) Capacity(id string) int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	tracker := t.trackers[id]
	if tracker == nil {
		return 1.0
	}
	return tracker.Capacity()
}

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
