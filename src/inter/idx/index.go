package idx

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

type (
	// Event numeration.
	Event uint64

	// Txn numeration.
	Txn uint64

	// Block numeration.
	Block uint64
)

func (t Txn) Bytes() []byte {
	return common.IntToBytes(uint64(t))
}

func (b Block) Bytes() []byte {
	return common.IntToBytes(uint64(b))
}

func BytesToBlock(b []byte) Block {
	var res Block
	for i := 0; i < len(b); i++ {
		res += Block(b[i]) << uint(i*8)
	}
	return res
}
