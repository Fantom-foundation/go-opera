package poset

// SetCheckpoint save Checkpoint.
// Checkpoint is seldom read; so no cache.
func (s *Store) SetCheckpoint(cp *Checkpoint) {
	const key = "current"
	s.set(s.table.Checkpoint, []byte(key), cp)
}

// GetCheckpoint returns stored Checkpoint.
// State is seldom read; so no cache.
func (s *Store) GetCheckpoint() *Checkpoint {
	const key = "current"
	w, _ := s.get(s.table.Checkpoint, []byte(key), &Checkpoint{}).(*Checkpoint)
	return w
}
