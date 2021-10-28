package gossip

import (
	"bytes"
	"sort"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/go-opera/utils/concurrent"
)

type sortedLastEvent []byte

func (es *epochStore) getCachedLastEvents() (*concurrent.ValidatorEventsSet, bool) {
	cache := es.cache.LastEvents.Load()
	if cache != nil {
		return cache.(*concurrent.ValidatorEventsSet), true
	}
	return nil, false
}

func (es *epochStore) loadLastEvents() *concurrent.ValidatorEventsSet {
	res := make(map[idx.ValidatorID]hash.Event, 100)

	b, err := es.table.LastEvents.Get([]byte{})
	if err != nil {
		es.Log.Crit("Failed to get key-value", "err", err)
	}
	if b == nil {
		return concurrent.WrapValidatorEventsSet(res)
	}
	for i := 0; i < len(b); i += 32 + 4 {
		res[idx.BytesToValidatorID(b[i:i+4])] = hash.BytesToEvent(b[i+4 : i+4+32])
	}

	return concurrent.WrapValidatorEventsSet(res)
}

func (es *epochStore) GetLastEvents() *concurrent.ValidatorEventsSet {
	cached, ok := es.getCachedLastEvents()
	if ok {
		return cached
	}
	heads := es.loadLastEvents()
	if heads == nil {
		heads = &concurrent.ValidatorEventsSet{}
	}
	es.cache.LastEvents.Store(heads)
	return heads
}

func (es *epochStore) SetLastEvents(ids *concurrent.ValidatorEventsSet) {
	es.cache.LastEvents.Store(ids)
}

func (es *epochStore) FlushLastEvents() {
	lasts, ok := es.getCachedLastEvents()
	if !ok {
		return
	}

	// sort values for determinism
	sortedLastEvents := make([]sortedLastEvent, 0, len(lasts.Val))
	for vid, val := range lasts.Val {
		b := append(vid.Bytes(), val.Bytes()...)
		sortedLastEvents = append(sortedLastEvents, b)
	}
	sort.Slice(sortedLastEvents, func(i, j int) bool {
		a, b := sortedLastEvents[i], sortedLastEvents[j]
		return bytes.Compare(a, b) < 0
	})

	b := make([]byte, 0, len(sortedLastEvents)*(32+4))
	for _, head := range sortedLastEvents {
		b = append(b, head...)
	}

	if err := es.table.LastEvents.Put([]byte{}, b); err != nil {
		es.Log.Crit("Failed to put key-value", "err", err)
	}
}

// GetLastEvents returns latest connected epoch events from each validator
func (s *Store) GetLastEvents(epoch idx.Epoch) *concurrent.ValidatorEventsSet {
	es := s.getEpochStore(epoch)
	if es == nil {
		return nil
	}

	return es.GetLastEvents()
}

// GetLastEvent returns latest connected epoch event from specified validator
func (s *Store) GetLastEvent(epoch idx.Epoch, vid idx.ValidatorID) *hash.Event {
	es := s.getEpochStore(epoch)
	if es == nil {
		return nil
	}

	lasts := s.GetLastEvents(epoch)
	lasts.RLock()
	defer lasts.RUnlock()
	last, ok := lasts.Val[vid]
	if !ok {
		return nil
	}
	return &last
}

func (s *Store) SetLastEvents(epoch idx.Epoch, ids *concurrent.ValidatorEventsSet) {
	es := s.getEpochStore(epoch)
	if es == nil {
		return
	}

	es.SetLastEvents(ids)
}
