package app

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/inter/sfctype"
)

// SetSfcDelegator stores SfcDelegator
func (s *Store) SetSfcDelegator(address common.Address, v *sfctype.SfcDelegator) {
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

// ForEachSfcDelegator iterates all stored SfcDelegators
func (s *Store) ForEachSfcDelegator(do func(sfctype.SfcDelegatorAndAddr)) {
	it := s.table.Delegators.NewIterator()
	defer it.Release()
	s.forEachSfcDelegator(it, do)
}

func (s *Store) forEachSfcDelegator(it ethdb.Iterator, do func(sfctype.SfcDelegatorAndAddr)) {
	for it.Next() {
		delegator := &sfctype.SfcDelegator{}
		err := rlp.DecodeBytes(it.Value(), delegator)
		if err != nil {
			s.Log.Crit("Failed to decode rlp while iteration", "err", err)
		}

		addr := it.Key()[len(it.Key())-20:]
		do(sfctype.SfcDelegatorAndAddr{
			Addr:      common.BytesToAddress(addr),
			Delegator: delegator,
		})
	}
}

// GetSfcDelegator returns stored SfcDelegator
func (s *Store) GetSfcDelegator(address common.Address) *sfctype.SfcDelegator {
	// Get data from LRU cache first.
	if s.cache.Delegators != nil {
		if c, ok := s.cache.Delegators.Get(address); ok {
			if b, ok := c.(*sfctype.SfcDelegator); ok {
				return b
			}
		}
	}

	w, _ := s.get(s.table.Delegators, address.Bytes(), &sfctype.SfcDelegator{}).(*sfctype.SfcDelegator)

	// Add to LRU cache.
	if w != nil && s.cache.Delegators != nil {
		s.cache.Delegators.Add(address, w)
	}

	return w
}
