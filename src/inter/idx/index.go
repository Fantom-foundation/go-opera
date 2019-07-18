package idx

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

type (
	// SuperFrame numeration.
	SuperFrame uint64

	// Event numeration.
	Event uint64

	// Txn numeration.
	Txn uint64

	// Block numeration.
	Block uint64
)

func (sf SuperFrame) Bytes() []byte {
	return common.IntToBytes(uint64(sf))
}

func (e Event) Bytes() []byte {
	return common.IntToBytes(uint64(e))
}

func (t Txn) Bytes() []byte {
	return common.IntToBytes(uint64(t))
}

func (b Block) Bytes() []byte {
	return common.IntToBytes(uint64(b))
}

func BytesToEvent(b []byte) Event {
	var res Event
	for i := 0; i < len(b); i++ {
		res += Event(b[i]) << uint(i*8)
	}
	return res
}

func BytesToTxn(b []byte) Txn {
	var res Txn
	for i := 0; i < len(b); i++ {
		res += Txn(b[i]) << uint(i*8)
	}
	return res
}

func BytesToBlock(b []byte) Block {
	var res Block
	for i := 0; i < len(b); i++ {
		res += Block(b[i]) << uint(i*8)
	}
	return res
}
