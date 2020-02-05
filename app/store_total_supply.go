package app

import (
	"math/big"
)

// GetTotalSupply returns total supply
func (s *Store) GetTotalSupply() *big.Int {
	amountBytes, err := s.table.TotalSupply.Get([]byte("c"))
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if amountBytes == nil {
		return big.NewInt(0)
	}
	return new(big.Int).SetBytes(amountBytes)
}

// SetTotalSupply stores total supply
func (s *Store) SetTotalSupply(amount *big.Int) {
	err := s.table.TotalSupply.Put([]byte("c"), amount.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}
}
