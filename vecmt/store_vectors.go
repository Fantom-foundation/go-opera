package vecmt

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
)

func (vi *Index) getBytes(table kvdb.Store, id hash.Event) []byte {
	key := id.Bytes()
	b, err := table.Get(key)
	if err != nil {
		vi.crit(err)
	}
	return b
}

func (vi *Index) setBytes(table kvdb.Store, id hash.Event, b []byte) {
	key := id.Bytes()
	err := table.Put(key, b)
	if err != nil {
		vi.crit(err)
	}
}

// GetHighestBeforeTime reads the vector from DB
func (vi *Index) GetHighestBeforeTime(id hash.Event) *HighestBeforeTime {
	if bVal, okGet := vi.cache.HighestBeforeTime.Get(id); okGet {
		return bVal.(*HighestBeforeTime)
	}

	b := HighestBeforeTime(vi.getBytes(vi.table.HighestBeforeTime, id))
	if b == nil {
		return nil
	}
	vi.cache.HighestBeforeTime.Add(id, &b, uint(len(b)))
	return &b
}

// GetHighestBefore reads the vector from DB
func (vi *Index) GetHighestBefore(id hash.Event) *HighestBefore {
	return &HighestBefore{
		VSeq:  vi.Base.GetHighestBefore(id),
		VTime: vi.GetHighestBeforeTime(id),
	}
}

// SetHighestBeforeTime stores the vector into DB
func (vi *Index) SetHighestBeforeTime(id hash.Event, vec *HighestBeforeTime) {
	vi.setBytes(vi.table.HighestBeforeTime, id, *vec)

	vi.cache.HighestBeforeTime.Add(id, vec, uint(len(*vec)))
}

// SetHighestBefore stores the vectors into DB
func (vi *Index) SetHighestBefore(id hash.Event, vec *HighestBefore) {
	vi.Base.SetHighestBefore(id, vec.VSeq)
	vi.SetHighestBeforeTime(id, vec.VTime)
}
