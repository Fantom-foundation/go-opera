package vector

import (
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
)

func (vi *Index) getBytes(table kvdb.KeyValueStore, id hash.Event) []byte {
	key := id.Bytes()
	b, err := table.Get(key)
	if err != nil {
		vi.Log.Crit("Failed to get key-value", "err", err)
	}
	return b
}

func (vi *Index) setBytes(table kvdb.KeyValueStore, id hash.Event, b []byte) {
	key := id.Bytes()
	err := table.Put(key, b)
	if err != nil {
		vi.Log.Crit("Failed to put key-value", "err", err)
	}
}

// GetLowestAfterSeq reads the vector from DB
func (vi *Index) GetLowestAfterSeq(id hash.Event) LowestAfterSeq {
	bVal, okGet := vi.cache.LowestAfterSeq.Get(id)
	b, okType := bVal.(LowestAfterSeq)
	if !okGet || !okType || b == nil {
		b = vi.getBytes(vi.table.LowestAfterSeq, id)

		vi.cache.LowestAfterSeq.Add(id, b)
	}

	return b
}

// GetHighestBeforeSeq reads the vector from DB
func (vi *Index) GetHighestBeforeSeq(id hash.Event) HighestBeforeSeq {
	bVal, okGet := vi.cache.HighestBeforeSeq.Get(id)
	b, okType := bVal.(HighestBeforeSeq)
	if !okGet || !okType || b == nil {
		b = vi.getBytes(vi.table.HighestBeforeSeq, id)

		vi.cache.HighestBeforeSeq.Add(id, b)
	}

	return b
}

// GetHighestBeforeTime reads the vector from DB
func (vi *Index) GetHighestBeforeTime(id hash.Event) HighestBeforeTime {
	bVal, okGet := vi.cache.HighestBeforeTime.Get(id)
	b, okType := bVal.(HighestBeforeTime)
	if !okGet || !okType || b == nil {
		b = vi.getBytes(vi.table.HighestBeforeTime, id)

		vi.cache.HighestBeforeTime.Add(id, b)
	}

	return b
}

// SetLowestAfter stores the vector into DB
func (vi *Index) SetLowestAfter(id hash.Event, seq LowestAfterSeq) {
	vi.setBytes(vi.table.LowestAfterSeq, id, seq)

	vi.cache.LowestAfterSeq.Add(id, seq)
}

// SetHighestBefore stores the vectors into DB
func (vi *Index) SetHighestBefore(id hash.Event, seq HighestBeforeSeq, time HighestBeforeTime) {
	vi.setBytes(vi.table.HighestBeforeSeq, id, seq)
	vi.setBytes(vi.table.HighestBeforeTime, id, time)

	vi.cache.HighestBeforeSeq.Add(id, seq)
	vi.cache.HighestBeforeTime.Add(id, time)
}

// setEventBranchID stores the event's global branch ID
func (vi *Index) setEventBranchID(id hash.Event, branchID idx.Validator) {
	vi.setBytes(vi.table.EventBranch, id, branchID.Bytes())

	vi.cache.EventBranch.Add(id, branchID.Bytes())
}

// getEventBranchID reads the event's global branch ID
func (vi *Index) getEventBranchID(id hash.Event) idx.Validator {
	bVal, okGet := vi.cache.EventBranch.Get(id)
	b, okType := bVal.([]byte)
	if !okGet || !okType || b == nil {
		b = vi.getBytes(vi.table.EventBranch, id)
		if b == nil {
			vi.Log.Crit("Failed to read event's branch ID (inconsistent DB)")
			return 0
		}

		vi.cache.EventBranch.Add(id, b)
	}
	branchID := idx.BytesToValidator(b)
	return branchID
}
