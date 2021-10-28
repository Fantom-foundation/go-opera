package gossip

import (
	"bytes"
	"sort"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/go-opera/utils/concurrent"
)

type sortedLastBV []byte

func (s *Store) getCachedLastBVs() (*concurrent.ValidatorBlocksSet, bool) {
	cache := s.cache.LastBVs.Load()
	if cache != nil {
		return cache.(*concurrent.ValidatorBlocksSet), true
	}
	return nil, false
}

func (s *Store) loadLastBVs() *concurrent.ValidatorBlocksSet {
	res := make(map[idx.ValidatorID]idx.Block, 100)

	b, err := s.table.LlrLastBlockVotes.Get([]byte{})
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if b == nil {
		return concurrent.WrapValidatorBlocksSet(res)
	}
	for i := 0; i < len(b); i += 8 + 4 {
		res[idx.BytesToValidatorID(b[i:i+4])] = idx.BytesToBlock(b[i+4 : i+4+8])
	}

	return concurrent.WrapValidatorBlocksSet(res)
}

func (s *Store) GetLastBVs() *concurrent.ValidatorBlocksSet {
	cached, ok := s.getCachedLastBVs()
	if ok {
		return cached
	}
	heads := s.loadLastBVs()
	if heads == nil {
		heads = &concurrent.ValidatorBlocksSet{}
	}
	s.cache.LastBVs.Store(heads)
	return heads
}

func (s *Store) SetLastBVs(ids *concurrent.ValidatorBlocksSet) {
	s.cache.LastBVs.Store(ids)
}

func (s *Store) FlushLastBVs() {
	lasts, ok := s.getCachedLastBVs()
	if !ok {
		return
	}

	// sort values for determinism
	sortedLastBVs := make([]sortedLastBV, 0, len(lasts.Val))
	for vid, val := range lasts.Val {
		b := append(vid.Bytes(), val.Bytes()...)
		sortedLastBVs = append(sortedLastBVs, b)
	}
	sort.Slice(sortedLastBVs, func(i, j int) bool {
		a, b := sortedLastBVs[i], sortedLastBVs[j]
		return bytes.Compare(a, b) < 0
	})

	b := make([]byte, 0, len(sortedLastBVs)*(8+4))
	for _, head := range sortedLastBVs {
		b = append(b, head...)
	}

	if err := s.table.LlrLastBlockVotes.Put([]byte{}, b); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// GetLastBV returns latest connected LLR block votes from specified validator
func (s *Store) GetLastBV(vid idx.ValidatorID) *idx.Block {
	lasts := s.GetLastBVs()
	lasts.RLock()
	defer lasts.RUnlock()
	last, ok := lasts.Val[vid]
	if !ok {
		return nil
	}
	return &last
}
