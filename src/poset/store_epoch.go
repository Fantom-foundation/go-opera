package poset

// SetGenesis stores first super-frame.
func (s *Store) SetGenesis(sf *epoch) {
	s.setEpoch([]byte("genesis"), sf)
}

// GetGenesis returns stored first super-frame.
func (s *Store) GetGenesis() *epoch {
	return s.getEpoch([]byte("genesis"))
}

// SetEpoch stores super-frame.
func (s *Store) SetEpoch(sf *epoch) {
	s.setEpoch([]byte("current"), sf)
}

// GetEpoch returns stored super-frame.
func (s *Store) GetEpoch() *epoch {
	return s.getEpoch([]byte("current"))
}

func (s *Store) setEpoch(key []byte, sf *epoch) {
	s.set(s.table.Epochs, key, sf)
}

func (s *Store) getEpoch(key []byte) *epoch {
	w, exists := s.get(s.table.Epochs, key, &epoch{}).(*epoch)
	if !exists {
		return nil
	}
	return w
}
