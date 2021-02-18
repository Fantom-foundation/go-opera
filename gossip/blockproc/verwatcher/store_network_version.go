package verwatcher

import (
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
)

const (
	nvKey = "v"
	mvKey = "m"
)

// SetNetworkVersion stores network version.
func (s *Store) SetNetworkVersion(v uint64) {
	err := s.mainDB.Put([]byte(nvKey), bigendian.Uint64ToBytes(v))
	if err != nil {
		s.Log.Crit("Failed to put key", "err", err)
	}
}

// GetNetworkVersion returns stored network version.
func (s *Store) GetNetworkVersion() uint64 {
	valBytes, err := s.mainDB.Get([]byte(nvKey))
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	if valBytes == nil {
		return 0
	}

	return bigendian.BytesToUint64(valBytes)
}

// SetMissedVersion stores non-supported network upgrade.
func (s *Store) SetMissedVersion(v uint64) {
	err := s.mainDB.Put([]byte(mvKey), bigendian.Uint64ToBytes(v))
	if err != nil {
		s.Log.Crit("Failed to put key", "err", err)
	}
}

// GetMissedVersion returns stored non-supported network upgrade.
func (s *Store) GetMissedVersion() uint64 {
	valBytes, err := s.mainDB.Get([]byte(mvKey))
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	if valBytes == nil {
		return 0
	}

	return bigendian.BytesToUint64(valBytes)
}
