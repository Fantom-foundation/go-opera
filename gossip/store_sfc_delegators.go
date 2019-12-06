package gossip

import (
	"github.com/ethereum/go-ethereum/common"
)

// SetSfcDelegator stores SfcDelegator
func (s *Store) SetSfcDelegator(address common.Address, v *SfcDelegator) {
	s.set(s.table.Delegators, address.Bytes(), v)

	// Add to LRU cache.
	if s.cache.Delegators != nil {
		s.cache.Delegators.Add(address, v)
	}
}

// DelSfcDelegator deletes SfcDelegator
func (s *Store) DelSfcDelegator(address common.Address) {
	err := s.table.Delegators.Delete(address.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase delegator")
	}

	// Add to LRU cache.
	if s.cache.Delegators != nil {
		s.cache.Delegators.Remove(address)
	}
}

// GetSfcDelegator returns stored SfcDelegator
func (s *Store) GetSfcDelegator(address common.Address) *SfcDelegator {
	// Get data from LRU cache first.
	if s.cache.Delegators != nil {
		if c, ok := s.cache.Delegators.Get(address); ok {
			if b, ok := c.(*SfcDelegator); ok {
				return b
			}
		}
	}

	w, _ := s.get(s.table.Delegators, address.Bytes(), &SfcDelegator{}).(*SfcDelegator)

	// Add to LRU cache.
	if w != nil && s.cache.Delegators != nil {
		s.cache.Delegators.Add(address, w)
	}

	return w
}
