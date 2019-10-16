package inter

import (
	"math"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/common/littleendian"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

const (
	MaxUint24 = uint64(math.MaxUint8) * math.MaxUint16
	MaxUint40 = uint64(math.MaxUint8) * math.MaxUint32
	MaxUint48 = uint64(math.MaxUint16) * math.MaxUint32
	MaxUint56 = MaxUint48 * math.MaxUint8
)

func (e *EventHeaderData) MarshalBinary() ([]byte, error) {
	// Calculate size of constant sized fields
	length := 53 + common.AddressLength + 2*common.HashLength

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
	length += 4 + parentsCount*common.HashLength + 4 + extraCount

	buf := make([]byte, length*2, length*2)
	offset := 0

	// Simple types values
	e.marshalOptimizedUint32(&buf, &offset)
	e.marshalOptimizedUint64(&buf, &offset)

	// Fixed types []byte values
	setToBuffer(&buf, &offset, e.Creator.Bytes(), common.AddressLength)
	setToBuffer(&buf, &offset, e.PrevEpochHash.Bytes(), common.HashLength)
	setToBuffer(&buf, &offset, e.TxHash.Bytes(), common.HashLength)

	// boolean
	b := byte(0)
	if e.IsRoot {
		b = 1
	}
	setToBuffer(&buf, &offset, []byte{b}, 1)

	// Parents
	e.marshalDeduplicateParents(&buf, &offset)

	setToBuffer(&buf, &offset, littleendian.Int32ToBytes(uint32(extraCount)), 4)
	setToBuffer(&buf, &offset, e.Extra, extraCount)

	bufLimited := buf[0:offset]

	return bufLimited, nil
}

func (e *EventHeaderData) UnmarshalBinary(buf []byte) error {
	// Simple types values
	offset := 0

	e.unmarshalOptimizedUint32(&buf, &offset)
	e.unmarshalOptimizedUint64(&buf, &offset)

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

func (e *EventHeaderData) marshalOptimizedUint32(buf *[]byte, offset *int) {
	// Detect max value from 4 fields
	b1 := maxBytesForUint32(e.Version)
	b2 := maxBytesForUint32(uint32(e.Epoch))
	b3 := maxBytesForUint32(uint32(e.Seq))
	b4 := maxBytesForUint32(uint32(e.Frame))
	sizeByte := byte((b1 - 1) | ((b2 - 1) << 2) | ((b3 - 1) << 4) | ((b4 - 1) << 6))

	setToBuffer(buf, offset, []byte{sizeByte}, 1)
	setToBuffer(buf, offset, littleendian.Int32ToBytes(e.Version), int(b1))
	setToBuffer(buf, offset, littleendian.Int32ToBytes(uint32(e.Epoch)), int(b2))
	setToBuffer(buf, offset, littleendian.Int32ToBytes(uint32(e.Seq)), int(b3))
	setToBuffer(buf, offset, littleendian.Int32ToBytes(uint32(e.Frame)), int(b4))

	b1 = maxBytesForUint32(uint32(e.Lamport))
	sizeByte = byte(b1 - 1)
	setToBuffer(buf, offset, []byte{sizeByte}, 1)
	setToBuffer(buf, offset, littleendian.Int32ToBytes(uint32(e.Lamport)), int(b1))
}

func (e *EventHeaderData) marshalOptimizedUint64(buf *[]byte, offset *int) {
	b1 := maxBytesForUint64(e.GasPowerLeft)
	b2 := maxBytesForUint64(e.GasPowerUsed)
	sizeByte := byte((b1 - 1) | ((b2 - 1) << 4))

	setToBuffer(buf, offset, []byte{sizeByte}, 1)
	setToBuffer(buf, offset, littleendian.Int64ToBytes(e.GasPowerLeft), int(b1))
	setToBuffer(buf, offset, littleendian.Int64ToBytes(e.GasPowerUsed), int(b2))

	b1 = maxBytesForUint64(uint64(e.ClaimedTime))
	b2 = maxBytesForUint64(uint64(e.MedianTime))
	sizeByte = byte((b1 - 1) | ((b2 - 1) << 4))
	setToBuffer(buf, offset, []byte{sizeByte}, 1)
	setToBuffer(buf, offset, littleendian.Int64ToBytes(uint64(e.ClaimedTime)), int(b1))
	setToBuffer(buf, offset, littleendian.Int64ToBytes(uint64(e.MedianTime)), int(b2))
}

func (e *EventHeaderData) marshalDeduplicateParents(buf *[]byte, offset *int) {
	// Sliced values
	parentsCount := len(e.Parents)
	setToBuffer(buf, offset, littleendian.Int32ToBytes(uint32(parentsCount)), 4)

	// Save epoch from first Parents (assume - all Parens have equal epoch part)
	if parentsCount > 0 {
		setToBuffer(buf, offset, e.Parents[0].Bytes(), 4)
	}

	for _, ev := range e.Parents {
		setToBuffer(buf, offset, ev.Bytes()[4:common.HashLength], common.HashLength-4)
	}
}

func (e *EventHeaderData) unmarshalOptimizedUint32(buf *[]byte, offset *int) {
	b1, b2, b3, b4 := readBufferSizeByte4Values(buf, offset)

	uint32buf := make([]byte, 4, 4)

	e.Version = readBufferOptimizedUint32(buf, offset, &uint32buf, b1)
	e.Epoch = idx.Epoch(readBufferOptimizedUint32(buf, offset, &uint32buf, b2))
	e.Seq = idx.Event(readBufferOptimizedUint32(buf, offset, &uint32buf, b3))
	e.Frame = idx.Frame(readBufferOptimizedUint32(buf, offset, &uint32buf, b4))

	b1, _, _, _ = readBufferSizeByte4Values(buf, offset)

	e.Lamport = idx.Lamport(readBufferOptimizedUint32(buf, offset, &uint32buf, b1))
}

func (e *EventHeaderData) unmarshalOptimizedUint64(buf *[]byte, offset *int) {
	b1, b2 := readBufferSizeByte2Values(buf, offset)

	uint64buf := make([]byte, 8, 8)

	e.GasPowerLeft = readBufferOptimizedUint64(buf, offset, &uint64buf, b1)
	e.GasPowerUsed = readBufferOptimizedUint64(buf, offset, &uint64buf, b2)

	b1, b2 = readBufferSizeByte2Values(buf, offset)

	e.ClaimedTime = Timestamp(readBufferOptimizedUint64(buf, offset, &uint64buf, b1))
	e.MedianTime = Timestamp(readBufferOptimizedUint64(buf, offset, &uint64buf, b2))
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
	if t <= math.MaxUint8 {
		return 1
	}
	if t <= math.MaxUint16 {
		return 2
	}
	if t <= MaxUint24 {
		return 3
	}
	if t <= math.MaxUint32 {
		return 4
	}
	if t <= MaxUint40 {
		return 5
	}
	if t <= MaxUint48 {
		return 6
	}
	if t <= MaxUint56 {
		return 7
	}
	return 8
}

func setToBuffer(buf *[]byte, offset *int, data []byte, size int) {
	copy((*buf)[*offset:*offset+size], data)
	*offset += size
}

func readBufferSizeByte4Values(buf *[]byte, offset *int) (b1 int, b2 int, b3 int, b4 int) {
	sizeByte := (*buf)[*offset]
	b1 = int(sizeByte&3 + 1)
	b2 = int((sizeByte>>2)&3 + 1)
	b3 = int((sizeByte>>4)&3 + 1)
	b4 = int((sizeByte>>6)&3 + 1)
	*offset++

	return
}

func readBufferSizeByte2Values(buf *[]byte, offset *int) (b1 int, b2 int) {
	sizeByte := (*buf)[*offset]
	b1 = int(sizeByte&7 + 1)
	b2 = int((sizeByte>>4)&7 + 1)
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
