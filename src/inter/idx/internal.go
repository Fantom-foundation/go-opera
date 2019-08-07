package idx

import (
	"github.com/Fantom-foundation/go-lachesis/src/common/bigendian"
)

type (
	// Member numeration.
	Member uint32
)

// Bytes gets the byte representation of the index.
func (m Member) Bytes() []byte {
	return bigendian.Int32ToBytes(uint32(m))
}

// BytesToMember converts bytes to member index.
func BytesToMember(b []byte) Member {
	return Member(bigendian.BytesToInt32(b))
}
