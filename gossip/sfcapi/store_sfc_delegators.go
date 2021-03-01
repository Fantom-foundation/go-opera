package sfcapi

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
)

// SetSfcDelegation stores SfcDelegation
func (s *Store) SetSfcDelegation(id DelegationID, v *SfcDelegation) {
	s.rlp.Set(s.table.Delegations, id.Bytes(), v)
}

// DelSfcDelegation deletes SfcDelegation
func (s *Store) DelSfcDelegation(id DelegationID) {
	err := s.table.Delegations.Delete(id.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase delegation")
	}
}

// ForEachSfcDelegation iterates all stored SfcDelegations
func (s *Store) ForEachSfcDelegation(do func(SfcDelegationAndID)) {
	it := s.table.Delegations.NewIterator(nil, nil)
	defer it.Release()
	s.forEachSfcDelegation(it, func(id SfcDelegationAndID) bool {
		do(id)
		return true
	})
}

// GetSfcDelegationsByAddr returns a lsit of delegations by address
func (s *Store) GetSfcDelegationsByAddr(addr common.Address, limit int) []SfcDelegationAndID {
	it := s.table.Delegations.NewIterator(addr.Bytes(), nil)
	defer it.Release()
	res := make([]SfcDelegationAndID, 0, limit)
	s.forEachSfcDelegation(it, func(id SfcDelegationAndID) bool {
		if limit == 0 {
			return false
		}
		limit--
		res = append(res, id)
		return true
	})
	return res
}

func (s *Store) forEachSfcDelegation(it ethdb.Iterator, do func(SfcDelegationAndID) bool) {
	_continue := true
	for _continue && it.Next() {
		delegation := &SfcDelegation{}
		err := rlp.DecodeBytes(it.Value(), delegation)
		if err != nil {
			s.Log.Crit("Failed to decode rlp while iteration", "err", err)
		}

		addr := it.Key()[len(it.Key())-DelegationIDSize:]
		_continue = do(SfcDelegationAndID{
			ID:         BytesToDelegationID(addr),
			Delegation: delegation,
		})
	}
}

// GetSfcDelegation returns stored SfcDelegation
func (s *Store) GetSfcDelegation(id DelegationID) *SfcDelegation {
	w, _ := s.rlp.Get(s.table.Delegations, id.Bytes(), &SfcDelegation{}).(*SfcDelegation)

	return w
}
