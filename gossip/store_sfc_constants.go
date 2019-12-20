package gossip

import "github.com/Fantom-foundation/go-lachesis/inter/idx"

// SetSfcConstants stores SfcConstants
func (s *Store) SetSfcConstants(epoch idx.Epoch, constants SfcConstants) {
	s.set(s.table.Delegators, epoch.Bytes(), constants)
}

// GetSfcConstants returns stored SfcConstants
func (s *Store) GetSfcConstants(epoch idx.Epoch) SfcConstants {
	w, _ := s.get(s.table.Delegators, epoch.Bytes(), &SfcConstants{}).(*SfcConstants)

	if w == nil {
		w = &SfcConstants{}
	}

	return *w
}
