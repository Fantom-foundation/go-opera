package poset

// SetCheckpoint save Checkpoint.
// Checkpoint is seldom read; so no cache.
func (s *Store) SetCheckpoint(cp *Checkpoint) {
	const key = "c"
	s.set(s.table.Checkpoint, []byte(key), cp)
}

// GetCheckpoint returns stored Checkpoint.
// State is seldom read; so no cache.
func (s *Store) GetCheckpoint() *Checkpoint {
	const key = "c"
	w, _ := s.get(s.table.Checkpoint, []byte(key), &Checkpoint{}).(*Checkpoint)
	return w
}
