package app

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/inter/sfctype"
)

// SetSfcDelegation stores SfcDelegation
func (s *Store) SetSfcDelegation(id sfctype.DelegationID, v *sfctype.SfcDelegation) {
	s.set(s.table.Delegations, id.Bytes(), v)

	// Add to LRU cache.
	if s.cache.Delegations != nil {
		s.cache.Delegations.Add(id, v)
	}
}

// DelSfcDelegation deletes SfcDelegation
func (s *Store) DelSfcDelegation(id sfctype.DelegationID) {
	err := s.table.Delegations.Delete(id.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase delegation")
	}

	// Add to LRU cache.
	if s.cache.Delegations != nil {
		s.cache.Delegations.Remove(id)
	}
}

// ForEachSfcDelegation iterates all stored SfcDelegations
func (s *Store) ForEachSfcDelegation(do func(sfctype.SfcDelegationAndID)) {
	it := s.table.Delegations.NewIterator()
	defer it.Release()
	s.forEachSfcDelegation(it, func(id sfctype.SfcDelegationAndID) bool {
		do(id)
		return true
	})
}

// GetSfcDelegationsByAddr returns a lsit of delegations by address
func (s *Store) GetSfcDelegationsByAddr(addr common.Address, limit int) []sfctype.SfcDelegationAndID {
	it := s.table.Delegations.NewIteratorWithPrefix(addr.Bytes())
	defer it.Release()
	res := make([]sfctype.SfcDelegationAndID, 0, limit)
	s.forEachSfcDelegation(it, func(id sfctype.SfcDelegationAndID) bool {
		res = append(res, id)
		limit -= 1
		return limit == 0
	})
	return res
}

func (s *Store) forEachSfcDelegation(it ethdb.Iterator, do func(sfctype.SfcDelegationAndID) bool) {
	_continue := true
	for _continue && it.Next() {
		delegation := &sfctype.SfcDelegation{}
		err := rlp.DecodeBytes(it.Value(), delegation)
		if err != nil {
			s.Log.Crit("Failed to decode rlp while iteration", "err", err)
		}

		addr := it.Key()[len(it.Key())-sfctype.DelegationIDSize:]
		_continue = do(sfctype.SfcDelegationAndID{
			ID:         sfctype.BytesToDelegationID(addr),
			Delegation: delegation,
		})
	}
}

// GetSfcDelegation returns stored SfcDelegation
func (s *Store) GetSfcDelegation(id sfctype.DelegationID) *sfctype.SfcDelegation {
	// Get data from LRU cache first.
	if s.cache.Delegations != nil {
		if c, ok := s.cache.Delegations.Get(id); ok {
			if b, ok := c.(*sfctype.SfcDelegation); ok {
				return b
			}
		}
	}

	w, _ := s.get(s.table.Delegations, id.Bytes(), &sfctype.SfcDelegation{}).(*sfctype.SfcDelegation)

	// Add to LRU cache.
	if w != nil && s.cache.Delegations != nil {
		s.cache.Delegations.Add(id, w)
	}

	return w
}
