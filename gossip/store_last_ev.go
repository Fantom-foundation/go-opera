package gossip

import (
	"bytes"
	"sort"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/go-opera/utils/concurrent"
)

type sortedLastEV []byte

func (s *Store) getCachedLastEV() (*concurrent.ValidatorEpochsSet, bool) {
	cache := s.cache.LastEV.Load()
	if cache != nil {
		return cache.(*concurrent.ValidatorEpochsSet), true
	}
	return nil, false
}

func (s *Store) loadLastEV() *concurrent.ValidatorEpochsSet {
	res := make(map[idx.ValidatorID]idx.Epoch, 100)

	b, err := s.table.LlrLastEpochVote.Get([]byte{})
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if b == nil {
		return concurrent.WrapValidatorEpochsSet(res)
	}
	for i := 0; i < len(b); i += 4 + 4 {
		res[idx.BytesToValidatorID(b[i:i+4])] = idx.BytesToEpoch(b[i+4 : i+4+4])
	}

	return concurrent.WrapValidatorEpochsSet(res)
}

func (s *Store) GetLastEVs() *concurrent.ValidatorEpochsSet {
	cached, ok := s.getCachedLastEV()
	if ok {
		return cached
	}
	heads := s.loadLastEV()
	if heads == nil {
		heads = &concurrent.ValidatorEpochsSet{}
	}
	s.cache.LastEV.Store(heads)
	return heads
}

func (s *Store) SetLastEVs(ids *concurrent.ValidatorEpochsSet) {
	s.cache.LastEV.Store(ids)
}

func (s *Store) FlushLastEV() {
	lasts, ok := s.getCachedLastEV()
	if !ok {
		return
	}

	// sort values for determinism
	sortedLastEV := make([]sortedLastEV, 0, len(lasts.Val))
	for vid, val := range lasts.Val {
		b := append(vid.Bytes(), val.Bytes()...)
		sortedLastEV = append(sortedLastEV, b)
	}
	sort.Slice(sortedLastEV, func(i, j int) bool {
		a, b := sortedLastEV[i], sortedLastEV[j]
		return bytes.Compare(a, b) < 0
	})

	b := make([]byte, 0, len(sortedLastEV)*(4+4))
	for _, head := range sortedLastEV {
		b = append(b, head...)
	}

	if err := s.table.LlrLastEpochVote.Put([]byte{}, b); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// GetLastEV returns latest connected LLR epoch vote from specified validator
func (s *Store) GetLastEV(vid idx.ValidatorID) *idx.Epoch {
	lasts := s.GetLastEVs()
	lasts.RLock()
	defer lasts.RUnlock()
	last, ok := lasts.Val[vid]
	if !ok {
		return nil
	}
	return &last
}
