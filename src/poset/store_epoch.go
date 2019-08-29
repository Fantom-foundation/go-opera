package poset

// SetGenesis stores first epoch.
func (s *Store) SetGenesis(e *epoch) {
	s.setEpoch([]byte("genesis"), e)
}

// GetGenesis returns stored first epoch.
func (s *Store) GetGenesis() *epoch {
	return s.getEpoch([]byte("genesis"))
}

// SetEpoch stores epoch.
func (s *Store) SetEpoch(e *epoch) {
	s.setEpoch([]byte("current"), e)
}

// GetEpoch returns stored epoch.
func (s *Store) GetEpoch() *epoch {
	return s.getEpoch([]byte("current"))
}

func (s *Store) setEpoch(key []byte, e *epoch) {
	s.set(s.table.Epochs, key, e)
}

func (s *Store) getEpoch(key []byte) *epoch {
	w, exists := s.get(s.table.Epochs, key, &epoch{}).(*epoch)
	if !exists {
		return nil
	}
	return w
}
