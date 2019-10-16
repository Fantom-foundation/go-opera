package inter

import (
	"math"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/common/littleendian"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

const (
	EventHeaderFixedDataSize = 53
	SerializedCounterSize = 4
	SerializedEpochSize = 4
)

type FastBuffer struct {
	buf *[]byte
	offset     	int
}

func NewFastBuffer(buf *[]byte) *FastBuffer {
	b := FastBuffer{
		buf:        buf,
		offset:     0,
	}

	return &b
}

func (b *FastBuffer) Write(src []byte) {
	size := len(src)
	b.WriteLen(src, size)
}

func (b *FastBuffer) WriteLen(src []byte, size int) {
	copy((*b.buf)[b.offset:b.offset+ size], src)
	b.offset += size
}

func (b *FastBuffer) Read(size int) (result []byte) {
	result = (*b.buf)[b.offset:b.offset+size]
	b.offset += size
	return
}

func (b *FastBuffer) Bytes() []byte {
	return (*b.buf)[0:b.offset]
}

func (e *EventHeaderData) MarshalBinary() ([]byte, error) {
	// Calculate size of constant sized fields
	length := EventHeaderFixedDataSize + common.AddressLength + 2*common.HashLength

	// Calculate sizes of slice fields
	parentsCount := 0
	if e.Parents != nil {
		parentsCount = len(e.Parents)
	}

	extraCount := 0
	if e.Extra != nil {
		extraCount = len(e.Extra)
	}

	// Full length with data about sizes of slices
	length += SerializedCounterSize + parentsCount*common.HashLength + SerializedCounterSize + extraCount

	bytesBuf := make([]byte, length*2, length*2)
	buf := NewFastBuffer(&bytesBuf)

	// Simple types values
	e.marshalUint32ToPacked(buf)
	e.marshalUint64ToPacked(buf)

	// Fixed types []byte values
	buf.Write(e.Creator.Bytes())
	buf.Write(e.PrevEpochHash.Bytes())
	buf.Write(e.TxHash.Bytes())

	// boolean
	b := byte(0)
	if e.IsRoot {
		b = 1
	}
	buf.Write([]byte{b})

	// Parents
	e.marshalDeduplicateParents(buf)

	buf.Write(littleendian.Int32ToBytes(uint32(extraCount))[0:SerializedCounterSize])
	buf.Write(e.Extra)

	return buf.Bytes(), nil
}

func (e *EventHeaderData) UnmarshalBinary(src []byte) error {
	// Simple types values
	buf := NewFastBuffer(&src)

	e.unmarshalPackedToUint32(buf)
	e.unmarshalPackedToUint64(buf)

	// Fixed types []byte values
	e.Creator.SetBytes(buf.Read(common.AddressLength))
	e.PrevEpochHash.SetBytes(buf.Read(common.HashLength))
	e.TxHash.SetBytes(buf.Read(common.HashLength))

	// Boolean
	e.IsRoot = readBufferBool(buf)

	// Sliced values
	e.unmarshalDeduplicateParents(buf)

	extraCount := readBufferUint32(buf)
	e.Extra = buf.Read(int(extraCount))

	return nil
}

func (e *EventHeaderData) marshalUint32ToPacked(buf *FastBuffer) {
	// Detect max value from 4 fields
	v1size := maxBytesForUint32(e.Version)
	v2size := maxBytesForUint32(uint32(e.Epoch))
	v3size := maxBytesForUint32(uint32(e.Seq))
	v4size := maxBytesForUint32(uint32(e.Frame))
	sizeByte := byte((v1size - 1) | ((v2size - 1) << 2) | ((v3size - 1) << 4) | ((v4size - 1) << 6))

	buf.Write([]byte{sizeByte})
	buf.Write(littleendian.Int32ToBytes(e.Version)[0:v1size])
	buf.Write(littleendian.Int32ToBytes(uint32(e.Epoch))[0:v2size])
	buf.Write(littleendian.Int32ToBytes(uint32(e.Seq))[0:v3size])
	buf.Write(littleendian.Int32ToBytes(uint32(e.Frame))[0:v4size])

	v1size = maxBytesForUint32(uint32(e.Lamport))
	sizeByte = byte(v1size - 1)
	buf.Write([]byte{sizeByte})
	buf.Write(littleendian.Int32ToBytes(uint32(e.Lamport))[0:v1size])
}

func (e *EventHeaderData) marshalUint64ToPacked(buf *FastBuffer) {
	v1size := maxBytesForUint64(e.GasPowerLeft)
	v2size := maxBytesForUint64(e.GasPowerUsed)
	sizeByte := byte((v1size - 1) | ((v2size - 1) << 4))

	buf.Write([]byte{sizeByte})
	buf.Write(littleendian.Int64ToBytes(e.GasPowerLeft)[0:v1size])
	buf.Write(littleendian.Int64ToBytes(e.GasPowerUsed)[0:v2size])

	v1size = maxBytesForUint64(uint64(e.ClaimedTime))
	v2size = maxBytesForUint64(uint64(e.MedianTime))
	sizeByte = byte((v1size - 1) | ((v2size - 1) << 4))
	buf.Write([]byte{sizeByte})
	buf.Write(littleendian.Int64ToBytes(uint64(e.ClaimedTime))[0:v1size])
	buf.Write(littleendian.Int64ToBytes(uint64(e.MedianTime))[0:v2size])
}

func (e *EventHeaderData) marshalDeduplicateParents(buf *FastBuffer) {
	// Sliced values
	parentsCount := len(e.Parents)
	buf.Write(littleendian.Int32ToBytes(uint32(parentsCount))[0:SerializedCounterSize])

	// Save epoch from first Parents (assume - all Parens have equal epoch part)
	if parentsCount > 0 {
		buf.Write(e.Parents[0].Bytes()[0:SerializedEpochSize])
	}

	for _, ev := range e.Parents {
		buf.Write(ev.Bytes()[4:common.HashLength])
	}
}

func (e *EventHeaderData) unmarshalPackedToUint32(buf *FastBuffer) {
	v1size, v2size, v3size, v4size := readBufferSizeByte4Values(buf)

	e.Version = readBufferPackedUint32(buf, v1size)
	e.Epoch = idx.Epoch(readBufferPackedUint32(buf, v2size))
	e.Seq = idx.Event(readBufferPackedUint32(buf, v3size))
	e.Frame = idx.Frame(readBufferPackedUint32(buf, v4size))

	v1size, _, _, _ = readBufferSizeByte4Values(buf)

	e.Lamport = idx.Lamport(readBufferPackedUint32(buf, v1size))
}

func (e *EventHeaderData) unmarshalPackedToUint64(buf *FastBuffer) {
	v1size, v2size := readBufferSizeByte2Values(buf)

	e.GasPowerLeft = readBufferPackedUint64(buf, v1size)
	e.GasPowerUsed = readBufferPackedUint64(buf, v2size)

	v1size, v2size = readBufferSizeByte2Values(buf)

	e.ClaimedTime = Timestamp(readBufferPackedUint64(buf, v1size))
	e.MedianTime = Timestamp(readBufferPackedUint64(buf, v2size))
}

func (e *EventHeaderData) unmarshalDeduplicateParents(buf *FastBuffer) {
	parentsCount := readBufferUint32(buf)

	evBuf := make([]byte, common.HashLength)
	if parentsCount > 0 {
		// Read epoch for all Parents
		copy(evBuf[0:4], buf.Read(4))
	}

	e.Parents = make(hash.Events, parentsCount, parentsCount)
	for i := 0; i < int(parentsCount); i++ {
		ev := hash.Event{}

		copy(evBuf[4:common.HashLength], buf.Read(common.HashLength-4))
		ev.SetBytes(evBuf)

		e.Parents[i] = ev
	}
}

func maxBytesForUint32(t uint32) uint {
	return maxBytesForUint64(uint64(t))
}

func maxBytesForUint64(t uint64) uint {
	mask := uint64(math.MaxUint64)
	for i := uint(1); i < 8; i++ {
		mask = mask << 8
		if mask & t == 0 {
			return i
		}
	}
	return 8
}

func readBufferSizeByte4Values(buf *FastBuffer) (v1size int, v2size int, v3size int, v4size int) {
	sizeByte := buf.Read(1)[0]
	v1size = int(sizeByte&3 + 1)
	v2size = int((sizeByte>>2)&3 + 1)
	v3size = int((sizeByte>>4)&3 + 1)
	v4size = int((sizeByte>>6)&3 + 1)

	return
}

func readBufferSizeByte2Values(buf *FastBuffer) (v1size int, v2size int) {
	sizeByte := buf.Read(1)[0]
	v1size = int(sizeByte&7 + 1)
	v2size = int((sizeByte>>4)&7 + 1)

	return
}

func readBufferPackedUint32(buf *FastBuffer, size int) uint32 {
	intBuf := []byte{0, 0, 0, 0}
	copy(intBuf, buf.Read(size))
	return littleendian.BytesToInt32(intBuf)
}

func readBufferPackedUint64(buf *FastBuffer, size int) uint64 {
	intBuf := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	copy(intBuf, buf.Read(size))
	return littleendian.BytesToInt64(intBuf)
}

func readBufferBool(buf *FastBuffer) bool {
	return buf.Read(1)[0] != byte(0)
}

func readBufferUint32(buf *FastBuffer) (data uint32) {
	return littleendian.BytesToInt32(buf.Read(4))
}
