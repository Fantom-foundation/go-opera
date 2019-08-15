package posposet

// SetCheckpoint save checkpoint.
// Checkpoint is seldom read; so no cache.
func (s *Store) SetCheckpoint(cp *checkpoint) {
	const key = "current"
	s.set(s.table.Checkpoint, []byte(key), cp)

}

// GetCheckpoint returns stored checkpoint.
// State is seldom read; so no cache.
func (s *Store) GetCheckpoint() *checkpoint {
	const key = "current"
	w, _ := s.get(s.table.Checkpoint, []byte(key), &checkpoint{}).(*checkpoint)
	return w
}
