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
	return vi.getBytes(vi.table.LowestAfterSeq, id)
}

// GetHighestBeforeSeq reads the vector from DB
func (vi *Index) GetHighestBeforeSeq(id hash.Event) HighestBeforeSeq {
	return vi.getBytes(vi.table.HighestBeforeSeq, id)
}

// GetHighestBeforeTime reads the vector from DB
func (vi *Index) GetHighestBeforeTime(id hash.Event) HighestBeforeTime {
	return vi.getBytes(vi.table.HighestBeforeTime, id)
}

// SetLowestAfter stores the vector into DB
func (vi *Index) SetLowestAfter(id hash.Event, seq LowestAfterSeq) {
	vi.setBytes(vi.table.LowestAfterSeq, id, seq)
}

// SetHighestBefore stores the vectors into DB
func (vi *Index) SetHighestBefore(id hash.Event, seq HighestBeforeSeq, time HighestBeforeTime) {
	vi.setBytes(vi.table.HighestBeforeSeq, id, seq)
	vi.setBytes(vi.table.HighestBeforeTime, id, time)
}

// setEventBranchID stores the event's global branch ID
func (vi *Index) setEventBranchID(id hash.Event, branchID idx.Validator) {
	vi.setBytes(vi.table.EventBranch, id, branchID.Bytes())
}

// getEventBranchID reads the event's global branch ID
func (vi *Index) getEventBranchID(id hash.Event) idx.Validator {
	b := vi.getBytes(vi.table.EventBranch, id)
	if b == nil {
		vi.Log.Crit("Failed to read event's branch ID (inconsistent DB)")
		return 0
	}
	branchID := idx.BytesToValidator(b)
	return branchID
}
