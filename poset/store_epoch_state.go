package poset

// SetGenesis stores first epoch.
func (s *Store) SetGenesis(e *epochState) {
	s.setEpoch([]byte("genesis"), e)
}

// GetGenesis returns stored first epoch.
func (s *Store) GetGenesis() *epochState {
	return s.getEpoch([]byte("genesis"))
}

// SetEpoch stores epoch.
func (s *Store) SetEpoch(e *epochState) {
	s.setEpoch([]byte("current"), e)
}

// GetEpoch returns stored epoch.
func (s *Store) GetEpoch() *epochState {
	return s.getEpoch([]byte("current"))
}

func (s *Store) setEpoch(key []byte, e *epochState) {
	s.set(s.table.Epochs, key, e)
}

func (s *Store) getEpoch(key []byte) *epochState {
	w, exists := s.get(s.table.Epochs, key, &epochState{}).(*epochState)
	if !exists {
		return nil
	}
	return w
}
