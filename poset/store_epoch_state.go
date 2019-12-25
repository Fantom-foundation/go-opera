package poset

import "github.com/ethereum/go-ethereum/common"

// SetGenesis stores first epoch.
func (s *Store) SetGenesis(e *EpochState) {
	// update cache
	s.cache.GenesisHash = nil

	s.setEpoch([]byte("g"), e)
}

// GetGenesis returns stored first epoch.
func (s *Store) GetGenesis() *EpochState {
	return s.getEpoch([]byte("g"))
}

// GetGenesisHash returns PrevEpochHash of first epoch.
func (s *Store) GetGenesisHash() common.Hash {
	if s.cache.GenesisHash != nil {
		return *s.cache.GenesisHash
	}

	epoch := s.GetGenesis()
	if epoch == nil {
		s.Log.Crit("No genesis found")
	}
	h := epoch.PrevEpoch.Hash()

	// update cache
	s.cache.GenesisHash = &h

	return h
}

// SetEpoch stores epoch.
func (s *Store) SetEpoch(e *EpochState) {
	s.setEpoch([]byte("c"), e)
}

// GetEpoch returns stored epoch.
func (s *Store) GetEpoch() *EpochState {
	return s.getEpoch([]byte("c"))
}

func (s *Store) setEpoch(key []byte, e *EpochState) {
	s.set(s.table.Epochs, key, e)
}

func (s *Store) getEpoch(key []byte) *EpochState {
	w, exists := s.get(s.table.Epochs, key, &EpochState{}).(*EpochState)
	if !exists {
		return nil
	}
	return w
}
