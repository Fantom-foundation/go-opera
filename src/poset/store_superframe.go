package poset

// SetGenesis stores first super-frame.
func (s *Store) SetGenesis(sf *superFrame) {
	s.setSuperFrame([]byte("genesis"), sf)
}

// GetGenesis returns stored first super-frame.
func (s *Store) GetGenesis() *superFrame {
	return s.getSuperFrame([]byte("genesis"))
}

// SetSuperFrame stores super-frame.
func (s *Store) SetSuperFrame(sf *superFrame) {
	s.setSuperFrame([]byte("current"), sf)
}

// GetSuperFrame returns stored super-frame.
func (s *Store) GetSuperFrame() *superFrame {
	return s.getSuperFrame([]byte("current"))
}

func (s *Store) setSuperFrame(key []byte, sf *superFrame) {
	s.set(s.table.SuperFrames, key, sf)
}

func (s *Store) getSuperFrame(key []byte) *superFrame {
	w, exists := s.get(s.table.SuperFrames, key, &superFrame{}).(*superFrame)
	if !exists {
		return nil
	}
	return w
}
