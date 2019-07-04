package idx

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

type (
	// SuperFrame numeration.
	SuperFrame uint64

	// Frame numeration.
	Frame uint32
)

func (sf SuperFrame) Bytes() []byte {
	return common.IntToBytes(uint64(sf))
}

func (f Frame) Bytes() []byte {
	return common.IntToBytes(uint64(f))
}

func BytesToFrame(b []byte) Frame {
	var res Frame
	for i := 0; i < len(b); i++ {
		res += Frame(b[i]) << uint(i*8)
	}
	return res
}
