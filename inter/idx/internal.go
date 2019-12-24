package idx

import (
	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
)

type (
	// Validator numeration.
	Validator uint32
)

// Bytes gets the byte representation of the index.
func (v Validator) Bytes() []byte {
	return bigendian.Int32ToBytes(uint32(v))
}

// BytesToValidator converts bytes to validator index.
func BytesToValidator(b []byte) Validator {
	return Validator(bigendian.BytesToInt32(b))
}

const (
	// ShortTermGas is short-window settings for gas power
	ShortTermGas = 0
	// LongTermGas is long-window settings for gas power
	LongTermGas = 1
)
