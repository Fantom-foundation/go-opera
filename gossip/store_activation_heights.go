package gossip

import (
	"github.com/Fantom-foundation/go-opera/opera"
)

func (s *Store) AddUpgradeHeight(h opera.UpgradeHeight) {
	orig := s.GetUpgradeHeights()
	// allocate new memory to avoid race condition in cache
	cp := make([]opera.UpgradeHeight, 0, len(orig)+1)
	cp = append(append(cp, orig...), h)

	s.rlp.Set(s.table.UpgradeHeights, []byte{}, cp)
	s.cache.UpgradeHeights.Store(cp)
}

func (s *Store) GetUpgradeHeights() []opera.UpgradeHeight {
	if v := s.cache.UpgradeHeights.Load(); v != nil {
		return v.([]opera.UpgradeHeight)
	}
	hh, ok := s.rlp.Get(s.table.UpgradeHeights, []byte{}, &[]opera.UpgradeHeight{}).(*[]opera.UpgradeHeight)
	if !ok {
		return []opera.UpgradeHeight{}
	}
	s.cache.UpgradeHeights.Store(*hh)
	return *hh
}
