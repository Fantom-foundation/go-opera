package common

import (
	"strconv"
	"sync"
)

// RollingIndex struct
type RollingIndex struct {
	name      string
	size      int
	lastIndex int64
	items     []interface{}
	locker    sync.RWMutex
}

// NewRollingIndex constructor
func NewRollingIndex(name string, size int) *RollingIndex {
	return &RollingIndex{
		name:      name,
		size:      size,
		items:     make([]interface{}, 0, 2*size),
		lastIndex: -1,
	}
}

// GetLastWindow getter
func (r *RollingIndex) GetLastWindow() (lastWindow []interface{}, lastIndex int64) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return r.items, r.lastIndex
}

// Get with skipIndex option
func (r *RollingIndex) Get(skipIndex int64) ([]interface{}, error) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	res := make([]interface{}, 0)

	if skipIndex > r.lastIndex {
		return res, nil
	}

	cachedItems := int64(len(r.items))
	//assume there are no gaps between indexes
	oldestCachedIndex := r.lastIndex - cachedItems + 1
	if skipIndex+1 < oldestCachedIndex {
		return res, NewStoreErr(r.name, TooLate, strconv.FormatInt(skipIndex, 10))
	}

	//index of 'skipped' in RollingIndex
	start := skipIndex - oldestCachedIndex + 1

	return r.items[start:], nil
}

// GetItem get item for a given index
func (r *RollingIndex) GetItem(index int64) (interface{}, error) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	items := int64(len(r.items))
	oldestCached := r.lastIndex - items + 1
	if index < oldestCached {
		return nil, NewStoreErr(r.name, TooLate, strconv.FormatInt(index, 10))
	}
	foundIndex := index - oldestCached
	if foundIndex >= items {
		return nil, NewStoreErr(r.name, KeyNotFound, strconv.FormatInt(index, 10))
	}
	return r.items[foundIndex], nil
}

// Set item for given index
func (r *RollingIndex) Set(item interface{}, index int64) error {
	r.locker.Lock()
	defer r.locker.Unlock()

	//only allow to set items with index <= lastIndex + 1
	//so that we may assume there are no gaps between items
	if 0 <= r.lastIndex && index > r.lastIndex+1 {
		return NewStoreErr(r.name, SkippedIndex, strconv.FormatInt(index, 10))
	}

	//adding a new item
	if r.lastIndex < 0 || (index == r.lastIndex+1) {
		if len(r.items) >= 2*r.size {
			r.roll()
		}
		r.items = append(r.items, item)
		r.lastIndex = index
		return nil
	}

	//replace and existing item
	//make sure index is also greater or equal than the oldest cached item's index
	cachedItems := int64(len(r.items))
	oldestCachedIndex := r.lastIndex - cachedItems + 1

	if index < oldestCachedIndex {
		return NewStoreErr(r.name, TooLate, strconv.FormatInt(index, 10))
	}

	//replacing existing item
	position := index - oldestCachedIndex //position of 'index' in RollingIndex
	r.items[position] = item

	return nil
}

func (r *RollingIndex) roll() {
	newList := make([]interface{}, 0, 2*r.size)
	newList = append(newList, r.items[r.size:]...)
	r.items = newList
}
