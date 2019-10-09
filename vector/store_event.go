package vector

import (
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
)

func (vi *Index) get(table kvdb.KeyValueStore, id hash.Event) []byte {
	key := id.Bytes()
	b, err := table.Get(key)
	if err != nil {
		vi.Log.Crit("Failed to get key-value", "err", err)
	}
	return b
}

func (vi *Index) set(table kvdb.KeyValueStore, id hash.Event, b []byte) {
	key := id.Bytes()
	err := table.Put(key, b)
	if err != nil {
		vi.Log.Crit("Failed to put key-value", "err", err)
	}
}

// GetLowestAfterSeq reads the vector from DB
func (vi *Index) GetLowestAfterSeq(id hash.Event) LowestAfterSeq {
	return vi.get(vi.table.LowestAfterSeq, id)
}

// GetHighestBeforeSeq reads the vector from DB
func (vi *Index) GetHighestBeforeSeq(id hash.Event) HighestBeforeSeq {
	return vi.get(vi.table.HighestBeforeSeq, id)
}

// GetHighestBeforeTime reads the vector from DB
func (vi *Index) GetHighestBeforeTime(id hash.Event) HighestBeforeTime {
	return vi.get(vi.table.HighestBeforeTime, id)
}

// SetLowestAfter stores the vector into DB
func (vi *Index) SetLowestAfter(id hash.Event, seq LowestAfterSeq) {
	vi.set(vi.table.LowestAfterSeq, id, seq)
}

// SetHighestBefore stores the vectors into DB
func (vi *Index) SetHighestBefore(id hash.Event, seq HighestBeforeSeq, time HighestBeforeTime) {
	vi.set(vi.table.HighestBeforeSeq, id, seq)
	vi.set(vi.table.HighestBeforeTime, id, time)
}
