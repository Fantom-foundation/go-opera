package poset

// SetGenesis stores first epoch.
func (s *Store) SetGenesis(e *EpochState) {
	s.setEpoch([]byte("genesis"), e)
}

// GetGenesis returns stored first epoch.
func (s *Store) GetGenesis() *EpochState {
	return s.getEpoch([]byte("genesis"))
}

// SetEpoch stores epoch.
func (s *Store) SetEpoch(e *EpochState) {
	s.setEpoch([]byte("current"), e)
}

// GetEpoch returns stored epoch.
func (s *Store) GetEpoch() *EpochState {
	return s.getEpoch([]byte("current"))
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
