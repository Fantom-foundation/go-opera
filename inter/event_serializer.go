package inter

/*
	Serializers for:
	- Event (RLP)
	- EventHeader (RLP)
	- EventHeaderData (Binary & RLP)
*/

import (
	"errors"
	"io"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/utils/fast"
)

type (
	eventHeaderType struct {
		EventHeaderData EventHeaderData
		Sig             []byte
	}

	eventType struct {
		EventHeader  eventHeaderType
		Transactions types.Transactions

		// caches
		size atomic.Value
	}
)

// EncodeRLP EventHeaderData RLP encode
func (e *EventHeaderData) EncodeRLP(w io.Writer) error {
	bytes, err := e.MarshalBinary()
	if err != nil {
		return err
	}

	err = rlp.Encode(w, &bytes)

	return err
}

// DecodeRLP EventHeaderData RLP decode
func (e *EventHeaderData) DecodeRLP(src *rlp.Stream) error {
	bytes, err := src.Bytes()
	if err != nil {
		return err
	}

	err = e.UnmarshalBinary(bytes)

	return err
}

// EncodeRLP EventHeader RLP encode
func (e *EventHeader) EncodeRLP(w io.Writer) error {
	eh := eventHeaderType{
		EventHeaderData: e.EventHeaderData,
		Sig:             e.Sig,
	}

	err := rlp.Encode(w, &eh)

	return err
}

// DecodeRLP EventHeader RLP decode
func (e *EventHeader) DecodeRLP(src *rlp.Stream) error {
	bytes, err := src.Raw()
	if err != nil {
		return err
	}

	eh := eventHeaderType{}
	err = rlp.DecodeBytes(bytes, &eh)
	if err != nil {
		return err
	}

	e.EventHeaderData = eh.EventHeaderData
	e.Sig = eh.Sig

	return nil
}


// EncodeRLP Event RLP encode
func (e *Event) EncodeRLP(w io.Writer) error {
	ev := eventType{
		EventHeader: eventHeaderType{
			EventHeaderData: e.EventHeaderData,
			Sig:             e.Sig,
		},
		Transactions: e.Transactions,
		size:         e.size,
	}

	err := rlp.Encode(w, &ev)

	return err
}

// DecodeRLP Event RLP decode
func (e *Event) DecodeRLP(src *rlp.Stream) error {
	bytes, err := src.Raw()
	if err != nil {
		return err
	}

	ev := eventType{}
	err = rlp.DecodeBytes(bytes, &ev)
	if err != nil {
		return err
	}

	e.EventHeader.EventHeaderData = ev.EventHeader.EventHeaderData
	e.EventHeader.Sig = ev.EventHeader.Sig
	e.Transactions = ev.Transactions

	return nil
}


// MarshalBinary EventHeaderData binary serialize
func (e *EventHeaderData) MarshalBinary() ([]byte, error) {
	parentsCount := 0
	if e.Parents != nil {
		parentsCount = len(e.Parents)
	}

	extraCount := 0
	if e.Extra != nil {
		extraCount = len(e.Extra)
	}

	fields32 := []uint32{
		e.Version,
		uint32(e.Epoch),
		uint32(e.Seq),
		uint32(e.Frame),
		uint32(e.Lamport),
		uint32(parentsCount),
		uint32(extraCount),
	}
	fields64 := []uint64{
		e.GasPowerLeft,
		e.GasPowerUsed,
		uint64(e.ClaimedTime),
		uint64(e.MedianTime),
	}
	fieldsBool := []bool{
		e.IsRoot,
	}

	// Calculate size max size for buf
	length := len(fields32)*4 + int(len(fields32)/4+1) + // int32 fields space + sizes header
		len(fields64)*8 + int(len(fields32)/4+1) + // int64 fields space + sizes header
		len(fieldsBool) +
		common.AddressLength + // Creator
		common.HashLength*2 + // PrevEpochHash, TxHash
		common.HashLength*parentsCount +
		extraCount

	bytesBuf := make([]byte, length)

	// Sizes in bytes, enough for save intX values, save to bit buffer

	// Calculate header size required for BitArray buffers
	header32Size := fast.BitArraySizeCalc(2, len(fields32))
	header64Size := fast.BitArraySizeCalc(4, len(fields64))
	headerBoolSize := fast.BitArraySizeCalc(1, len(fieldsBool))
	headerSize := header32Size + header64Size + headerBoolSize

	// Set buffers like slice from main buffer
	header32buf := bytesBuf[0:header32Size]
	header64buf := bytesBuf[header32Size : header32Size+header64Size]
	headerBoolBuf := bytesBuf[header32Size+header64Size : headerSize]

	// Create BitArrays for work with headers
	header32 := fast.NewBitArray(2, &header32buf)
	header64 := fast.NewBitArray(4, &header64buf)
	headerBool := fast.NewBitArray(1, &headerBoolBuf)

	// Separate slices for header and data buffer
	dataBuf := bytesBuf[headerSize:]
	buf := fast.NewBuffer(&dataBuf)

	for _, i32 := range fields32 {
		n := writeUint32Compact(buf, i32)
		header32.Push(n - 1)
	}

	for _, i64 := range fields64 {
		n := writeUint64Compact(buf, i64)
		header64.Push(n - 1)
	}

	for _, b := range fieldsBool {
		if b {
			headerBool.Push(1)
		} else {
			headerBool.Push(0)
		}
	}

	// Fixed types []byte values
	buf.Write(e.Creator.Bytes())
	buf.Write(e.PrevEpochHash.Bytes())
	buf.Write(e.TxHash.Bytes())

	// Write parents without Epoch (-4 bytes per parent for save)
	for _, parent := range e.Parents {
		buf.Write(parent.Bytes()[4:common.HashLength])
	}

	buf.Write(e.Extra)

	return bytesBuf[0 : headerSize+buf.BytesLen()], nil
}

// UnmarshalBinary EventHeaderData binary deserialize
func (e *EventHeaderData) UnmarshalBinary(src []byte) error {
	var parentCount uint32
	var extraCount uint32

	fields32 := []*uint32{
		&e.Version,
		(*uint32)(&e.Epoch),
		(*uint32)(&e.Seq),
		(*uint32)(&e.Frame),
		(*uint32)(&e.Lamport),
		&parentCount,
		&extraCount,
	}
	fields64 := []*uint64{
		&e.GasPowerLeft,
		&e.GasPowerUsed,
		(*uint64)(&e.ClaimedTime),
		(*uint64)(&e.MedianTime),
	}
	fieldsBool := []*bool{
		&e.IsRoot,
	}

	// Create buffers for read bits packed data about sizes of intX values
	header32Size := fast.BitArraySizeCalc(2, len(fields32))
	header64Size := fast.BitArraySizeCalc(4, len(fields64))
	headerBoolSize := fast.BitArraySizeCalc(1, len(fieldsBool))
	headerSize := header32Size + header64Size + headerBoolSize

	header32buf := src[0:header32Size]
	header64buf := src[header32Size : header32Size+header64Size]
	headerBoolBuf := src[header32Size+header64Size : headerSize]

	header32 := fast.NewBitArray(2, &header32buf)
	header64 := fast.NewBitArray(4, &header64buf)
	headerBool := fast.NewBitArray(1, &headerBoolBuf)

	// Buffer for data read
	dataBuf := src[headerSize:]
	buf := fast.NewBuffer(&dataBuf)

	var err error
	for _, f := range fields32 {
		n := header32.Pop()
		*f, err = readUint32Compact(buf, n+1)
		if err != nil {
			return err
		}
	}

	for _, f := range fields64 {
		n := header64.Pop()
		*f, err = readUint64Compact(buf, n+1)
		if err != nil {
			return err
		}
	}

	for _, f := range fieldsBool {
		*f = headerBool.Pop() != 0
	}

	// Fixed types []byte values
	e.Creator.SetBytes(buf.Read(common.AddressLength))
	e.PrevEpochHash.SetBytes(buf.Read(common.HashLength))
	e.TxHash.SetBytes(buf.Read(common.HashLength))

	// Read parents without epoch and set epoch from e.Epoch
	e.Parents = make(hash.Events, parentCount)
	for i := uint32(0); i < parentCount; i++ {
		copy(e.Parents[i][:4], e.Epoch.Bytes())
		copy(e.Parents[i][4:], buf.Read(common.HashLength-4))
	}

	e.Extra = buf.Read(int(extraCount))

	return nil
}

func writeUint32Compact(buf *fast.Buffer, v uint32) (bytes int) {
	for {
		_ = buf.WriteByte(byte(v))
		bytes++
		v = v >> 8
		if v == 0 {
			return
		}
	}
}

func writeUint64Compact(buf *fast.Buffer, v uint64) (bytes int) {
	for {
		_ = buf.WriteByte(byte(v))
		bytes++
		v = v >> 8
		if v == 0 {
			return
		}
	}
}

func readUint32Compact(buf *fast.Buffer, bytes int) (uint32, error) {
	var v uint32
	for i, b := range buf.Read(bytes) {
		v += uint32(b) << uint(8*i)
		if b == 0 && i > 0 {
			return v, errors.New("ambiguous int32 decoding")
		}
	}
	return v, nil
}

func readUint64Compact(buf *fast.Buffer, bytes int) (uint64, error) {
	var v uint64
	for i, b := range buf.Read(bytes) {
		v += uint64(b) << uint(8*i)
		if b == 0 && i > 0 {
			return v, errors.New("ambiguous int64 decoding")
		}
	}
	return v, nil
}
