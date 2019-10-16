package inter

import (
	"bytes"
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

	buf := bytes.NewBuffer(make([]byte, 0, length*2))

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

func (e *EventHeaderData) UnmarshalBinary(buf []byte) error {
	// Simple types values
	offset := 0

	e.unmarshalPackedToUint32(&buf, &offset)
	e.unmarshalPackedToUint64(&buf, &offset)

	// Fixed types []byte values
	e.Creator.SetBytes(readBufferBytes(&buf, &offset, common.AddressLength))
	e.PrevEpochHash.SetBytes(readBufferBytes(&buf, &offset, common.HashLength))
	e.TxHash.SetBytes(readBufferBytes(&buf, &offset, common.HashLength))

	// Boolean
	e.IsRoot = readBufferBool(&buf, &offset)

	// Sliced values
	e.unmarshalDeduplicateParents(&buf, &offset)

	extraCount := readBufferUint32(&buf, &offset)
	e.Extra = readBufferBytes(&buf, &offset, int(extraCount))

	return nil
}

func (e *EventHeaderData) marshalUint32ToPacked(buf *bytes.Buffer) {
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

func (e *EventHeaderData) marshalUint64ToPacked(buf *bytes.Buffer) {
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

func (e *EventHeaderData) marshalDeduplicateParents(buf *bytes.Buffer) {
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

func (e *EventHeaderData) unmarshalPackedToUint32(buf *[]byte, offset *int) {
	v1size, v2size, v3size, v4size := readBufferSizeByte4Values(buf, offset)

	uint32buf := make([]byte, 4, 4)

	e.Version = readBufferOptimizedUint32(buf, offset, &uint32buf, v1size)
	e.Epoch = idx.Epoch(readBufferOptimizedUint32(buf, offset, &uint32buf, v2size))
	e.Seq = idx.Event(readBufferOptimizedUint32(buf, offset, &uint32buf, v3size))
	e.Frame = idx.Frame(readBufferOptimizedUint32(buf, offset, &uint32buf, v4size))

	v1size, _, _, _ = readBufferSizeByte4Values(buf, offset)

	e.Lamport = idx.Lamport(readBufferOptimizedUint32(buf, offset, &uint32buf, v1size))
}

func (e *EventHeaderData) unmarshalPackedToUint64(buf *[]byte, offset *int) {
	v1size, v2size := readBufferSizeByte2Values(buf, offset)

	uint64buf := make([]byte, 8, 8)

	e.GasPowerLeft = readBufferOptimizedUint64(buf, offset, &uint64buf, v1size)
	e.GasPowerUsed = readBufferOptimizedUint64(buf, offset, &uint64buf, v2size)

	v1size, v2size = readBufferSizeByte2Values(buf, offset)

	e.ClaimedTime = Timestamp(readBufferOptimizedUint64(buf, offset, &uint64buf, v1size))
	e.MedianTime = Timestamp(readBufferOptimizedUint64(buf, offset, &uint64buf, v2size))
}

func (e *EventHeaderData) unmarshalDeduplicateParents(buf *[]byte, offset *int) {
	parentsCount := readBufferUint32(buf, offset)

	evBuf := make([]byte, common.HashLength)
	if parentsCount > 0 {
		// Read epoch for all Parents
		copy(evBuf[0:4], readBufferBytes(buf, offset, 4))
	}

	e.Parents = make(hash.Events, parentsCount, parentsCount)
	for i := 0; i < int(parentsCount); i++ {
		ev := hash.Event{}

		copy(evBuf[4:common.HashLength], readBufferBytes(buf, offset, common.HashLength-4))
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

func readBufferSizeByte4Values(buf *[]byte, offset *int) (v1size int, v2size int, v3size int, v4size int) {
	sizeByte := (*buf)[*offset]
	v1size = int(sizeByte&3 + 1)
	v2size = int((sizeByte>>2)&3 + 1)
	v3size = int((sizeByte>>4)&3 + 1)
	v4size = int((sizeByte>>6)&3 + 1)
	*offset++

	return
}

func readBufferSizeByte2Values(buf *[]byte, offset *int) (v1size int, v2size int) {
	sizeByte := (*buf)[*offset]
	v1size = int(sizeByte&7 + 1)
	v2size = int((sizeByte>>4)&7 + 1)
	*offset++

	return
}

func readBufferOptimizedUint32(buf *[]byte, offset *int, intBuf *[]byte, size int) uint32 {
	copy(*intBuf, []byte{0, 0, 0, 0})
	copy(*intBuf, (*buf)[*offset:*offset+size])
	*offset += size
	return littleendian.BytesToInt32(*intBuf)
}

func readBufferOptimizedUint64(buf *[]byte, offset *int, intBuf *[]byte, size int) uint64 {
	copy(*intBuf, []byte{0, 0, 0, 0, 0, 0, 0, 0})
	copy(*intBuf, (*buf)[*offset:*offset+size])
	*offset += size
	return littleendian.BytesToInt64(*intBuf)
}

func readBufferBytes(buf *[]byte, offset *int, size int) (data []byte) {
	data = (*buf)[*offset : *offset+size]
	*offset += size

	return
}

func readBufferBool(buf *[]byte, offset *int) (data bool) {
	data = (*buf)[*offset] != byte(0)
	*offset++

	return
}

func readBufferUint32(buf *[]byte, offset *int) (data uint32) {
	data = littleendian.BytesToInt32((*buf)[*offset : *offset+4])
	*offset += 4

	return
}
