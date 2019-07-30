package idx

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

type (
	// SuperFrame numeration.
	SuperFrame uint64

	// Event numeration.
	Event uint64

	// Member numeration.
	Member uint64

	// Txn numeration.
	Txn uint64

	// Block numeration.
	Block uint64

	// Lamport numeration.
	Lamport uint64
)

// Bytes gets the byte representation of the index.
func (sf SuperFrame) Bytes() []byte {
	return common.IntToBytes(uint64(sf))
}

// Bytes gets the byte representation of the index.
func (e Event) Bytes() []byte {
	return common.IntToBytes(uint64(e))
}

// Bytes gets the byte representation of the index.
func (m Member) Bytes() []byte {
	return common.IntToBytes(uint64(m))
}

// Bytes gets the byte representation of the index.
func (t Txn) Bytes() []byte {
	return common.IntToBytes(uint64(t))
}

// Bytes gets the byte representation of the index.
func (b Block) Bytes() []byte {
	return common.IntToBytes(uint64(b))
}

// Bytes gets the byte representation of the index.
func (l Lamport) Bytes() []byte {
	return common.IntToBytes(uint64(l))
}

// BytesToEvent converts bytes to event index.
func BytesToEvent(b []byte) Event {
	var res Event
	for i := 0; i < len(b); i++ {
		res += Event(b[i]) << uint(i*8)
	}
	return res
}

// BytesToMember converts bytes to member index.
func BytesToMember(b []byte) Member {
	var res Member
	for i := 0; i < len(b); i++ {
		res += Member(b[i]) << uint(i*8)
	}
	return res
}

// BytesToTxn converts bytes to transaction index.
func BytesToTxn(b []byte) Txn {
	var res Txn
	for i := 0; i < len(b); i++ {
		res += Txn(b[i]) << uint(i*8)
	}
	return res
}

// BytesToBlock converts bytes to block index.
func BytesToBlock(b []byte) Block {
	var res Block
	for i := 0; i < len(b); i++ {
		res += Block(b[i]) << uint(i*8)
	}
	return res
}

// BytesToBlock converts bytes to block index.
func BytesToLamport(b []byte) Lamport {
	var res Lamport
	for i := 0; i < len(b); i++ {
		res += Lamport(b[i]) << uint(i*8)
	}
	return res
}
