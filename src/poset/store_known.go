// +build test

package poset

// KnownEvents returns all known events
func KnownEvents(s Store) map[uint64]int64 {
	known := make(map[uint64]int64)
	participants, _ := s.Participants()
	participants.RLock()
	defer participants.RUnlock()
	for p, pid := range participants.ByPubKey {
		index := int64(-1)
		last, isRoot, err := s.LastEventFrom(p)
		if err == nil {
			if isRoot {
				root, err := s.GetRoot(p)
				if err != nil {
					index = root.SelfParent.Index
				}
			} else {
				lastEvent, err := s.GetEventBlock(last)
				if err == nil {
					index = lastEvent.Index()
				}
			}

		}
		known[pid.ID] = index
	}
	return known
}
