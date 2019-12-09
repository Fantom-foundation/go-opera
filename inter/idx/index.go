package idx

import (
	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
)

type (
	// Epoch numeration.
	Epoch uint32

	// Event numeration.
	Event uint32

	// Txn numeration.
	Txn uint32

	// Block numeration.
	Block uint64

	// Lamport numeration.
	Lamport uint32

	// Frame numeration.
	Frame uint32

	// Pack numeration.
	Pack uint32

	// StakerID numeration.
	StakerID uint32
)

// Bytes gets the byte representation of the index.
func (e Epoch) Bytes() []byte {
	return bigendian.Int32ToBytes(uint32(e))
}

// Bytes gets the byte representation of the index.
func (e Event) Bytes() []byte {
	return bigendian.Int32ToBytes(uint32(e))
}

// Bytes gets the byte representation of the index.
func (t Txn) Bytes() []byte {
	return bigendian.Int32ToBytes(uint32(t))
}

// Bytes gets the byte representation of the index.
func (b Block) Bytes() []byte {
	return bigendian.Int64ToBytes(uint64(b))
}

// Bytes gets the byte representation of the index.
func (l Lamport) Bytes() []byte {
	return bigendian.Int32ToBytes(uint32(l))
}

// Bytes gets the byte representation of the index.
func (p Pack) Bytes() []byte {
	return bigendian.Int32ToBytes(uint32(p))
}

// Bytes gets the byte representation of the index.
func (s StakerID) Bytes() []byte {
	return bigendian.Int32ToBytes(uint32(s))
}

// Bytes gets the byte representation of the index.
func (f Frame) Bytes() []byte {
	return bigendian.Int32ToBytes(uint32(f))
}

// BytesToEpoch converts bytes to epoch index.
func BytesToEpoch(b []byte) Epoch {
	return Epoch(bigendian.BytesToInt32(b))
}

// BytesToEvent converts bytes to event index.
func BytesToEvent(b []byte) Event {
	return Event(bigendian.BytesToInt32(b))
}

// BytesToTxn converts bytes to transaction index.
func BytesToTxn(b []byte) Txn {
	return Txn(bigendian.BytesToInt32(b))
}

// BytesToBlock converts bytes to block index.
func BytesToBlock(b []byte) Block {
	return Block(bigendian.BytesToInt64(b))
}

// BytesToLamport converts bytes to block index.
func BytesToLamport(b []byte) Lamport {
	return Lamport(bigendian.BytesToInt32(b))
}

// BytesToFrame converts bytes to block index.
func BytesToFrame(b []byte) Frame {
	return Frame(bigendian.BytesToInt32(b))
}

// BytesToPack converts bytes to block index.
func BytesToPack(b []byte) Pack {
	return Pack(bigendian.BytesToInt32(b))
}

// BytesToStakerID converts bytes to staker index.
func BytesToStakerID(b []byte) StakerID {
	return StakerID(bigendian.BytesToInt32(b))
}

// MaxLamport return max value
func MaxLamport(x, y Lamport) Lamport {
	if x > y {
		return x
	}
	return y
}
