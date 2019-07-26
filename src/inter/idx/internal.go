package idx

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

type (
	// Frame numeration.
	Frame uint32
)

// Bytes gets the byte representation of the index.
func (f Frame) Bytes() []byte {
	return common.IntToBytes(uint64(f))
}

// BytesToFrame converts bytes to frame index.
func BytesToFrame(b []byte) Frame {
	var res Frame
	for i := 0; i < len(b); i++ {
		res += Frame(b[i]) << uint(i*8)
	}
	return res
}
