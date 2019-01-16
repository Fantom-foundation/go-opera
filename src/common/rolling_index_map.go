package common

import (
	"fmt"
)

// RollingIndexMap struct
type RollingIndexMap struct {
	name    string
	size    int
	keys    []uint64
	mapping map[uint64]*RollingIndex
}

// NewRollingIndexMap constructor
func NewRollingIndexMap(name string, size int, keys []uint64) *RollingIndexMap {
	items := make(map[uint64]*RollingIndex)
	for _, key := range keys {
		items[key] = NewRollingIndex(fmt.Sprintf("%s[%d]", name, key), size)
	}
	return &RollingIndexMap{
		name:    name,
		size:    size,
		keys:    keys,
		mapping: items,
	}
}

// AddKey adds a key to map if it doesn't already exist, returns an error if it already exists
func (rim *RollingIndexMap) AddKey(key uint64) error {
	if _, ok := rim.mapping[key]; ok {
		return NewStoreErr(rim.name, KeyAlreadyExists, fmt.Sprintf("%d", key))
	}
	rim.keys = append(rim.keys, key)
	rim.mapping[key] = NewRollingIndex(fmt.Sprintf("%s[%d]", rim.name, key), rim.size)
	return nil
}

// Get return key items with index > skip
func (rim *RollingIndexMap) Get(key uint64, skipIndex int64) ([]interface{}, error) {
	items, ok := rim.mapping[key]
	if !ok {
		return nil, NewStoreErr(rim.name, KeyNotFound, fmt.Sprint(key))
	}

	cached, err := items.Get(skipIndex)
	if err != nil {
		return nil, err
	}

	return cached, nil
}

// GetItem returns the item in rolling index for given key
func (rim *RollingIndexMap) GetItem(key uint64, index int64) (interface{}, error) {
	return rim.mapping[key].GetItem(index)
}

// GetLast get last item for given key
func (rim *RollingIndexMap) GetLast(key uint64) (interface{}, error) {
	pe, ok := rim.mapping[key]
	if !ok {
		return nil, NewStoreErr(rim.name, KeyNotFound, fmt.Sprint(key))
	}
	cached, _ := pe.GetLastWindow()
	if len(cached) == 0 {
		return "", NewStoreErr(rim.name, Empty, "")
	}
	return cached[len(cached)-1], nil
}

// Set sets a given key for the index
func (rim *RollingIndexMap) Set(key uint64, item interface{}, index int64) error {
	items, ok := rim.mapping[key]
	if !ok {
		items = NewRollingIndex(fmt.Sprintf("%s[%d]", rim.name, key), rim.size)
		rim.mapping[key] = items
	}
	return items.Set(item, index)
}

// Known returns [key] => lastKnownIndex
func (rim *RollingIndexMap) Known() map[uint64]int64 {
	known := make(map[uint64]int64)
	for k, items := range rim.mapping {
		_, lastIndex := items.GetLastWindow()
		known[k] = lastIndex
	}
	return known
}

// Reset resets the map
func (rim *RollingIndexMap) Reset() error {
	items := make(map[uint64]*RollingIndex)
	for _, key := range rim.keys {
		items[key] = NewRollingIndex(fmt.Sprintf("%s[%d]", rim.name, key), rim.size)
	}
	rim.mapping = items
	return nil
}

// Import from another index
func (rim *RollingIndexMap) Import(other *RollingIndexMap) {
	for _, key := range other.keys {
		rim.mapping[key] = NewRollingIndex(fmt.Sprintf("%s[%d]", rim.name, key), rim.size)
		rim.mapping[key].lastIndex = other.mapping[key].lastIndex
		rim.mapping[key].items = other.mapping[key].items
		// copy(rim.mapping[key].items[:len(other.mapping[key].items)], other.mapping[key].items)
	}
}
