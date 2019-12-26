package inter

import (
	"errors"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/utils/fast"
)

var (
	ErrNonCanonicalEncoding = errors.New("Non canonical encoded event")
	ErrInvalidEncoding      = errors.New("Invalid encoded event")
)

type (
	eventMarshaling struct {
		EventHeader  eventHeaderMarshaling
		Transactions types.Transactions
	}

	eventHeaderMarshaling struct {
		EventHeaderData EventHeaderData
		Sig             []byte
	}
)

// EncodeRLP implements rlp.Encoder interface.
// Its goal to override custom encoder of promoted field EventHeaderData.
func (e *Event) EncodeRLP(w io.Writer) error {
	m := &eventMarshaling{
		EventHeader: eventHeaderMarshaling{
			EventHeaderData: e.EventHeaderData,
			Sig:             e.Sig,
		},
		Transactions: e.Transactions,
	}

	return rlp.Encode(w, m)
}

// DecodeRLP implements rlp.Decoder interface.
// Its goal to override custom decoder of promoted field EventHeaderData.
func (e *Event) DecodeRLP(src *rlp.Stream) error {
	m := &eventMarshaling{}
	if err := src.Decode(m); err != nil {
		return err
	}

	e.EventHeader.EventHeaderData = m.EventHeader.EventHeaderData
	e.EventHeader.Sig = m.EventHeader.Sig
	e.Transactions = m.Transactions

	return nil
}

// EncodeRLP implements rlp.Encoder interface.
// Its goal to override custom encoder of promoted field EventHeaderData.
func (e *EventHeader) EncodeRLP(w io.Writer) error {
	m := &eventHeaderMarshaling{
		EventHeaderData: e.EventHeaderData,
		Sig:             e.Sig,
	}

	return rlp.Encode(w, m)
}

// DecodeRLP implements rlp.Decoder interface.
// Its goal to override custom decoder of promoted field EventHeaderData.
func (e *EventHeader) DecodeRLP(src *rlp.Stream) error {
	m := &eventHeaderMarshaling{}
	if err := src.Decode(m); err != nil {
		return err
	}

	e.EventHeaderData = m.EventHeaderData
	e.Sig = m.Sig

	return nil
}

// EncodeRLP implements rlp.Encoder interface.
func (e *EventHeaderData) EncodeRLP(w io.Writer) error {
	bytes, err := e.MarshalBinary()
	if err != nil {
		return err
	}

	err = rlp.Encode(w, &bytes)

	return err
}

// DecodeRLP implements rlp.Decoder interface.
func (e *EventHeaderData) DecodeRLP(src *rlp.Stream) error {
	bytes, err := src.Bytes()
	if err != nil {
		return err
	}

	return e.UnmarshalBinary(bytes)
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (e *EventHeaderData) MarshalBinary() ([]byte, error) {
	isPrevEpochHashEmpty := (e.PrevEpochHash == hash.Zero)
	isTxHashEmpty := (e.TxHash == EmptyTxHash)

	fields64 := []uint64{
		e.GasPowerLeft.Gas[0],
		e.GasPowerLeft.Gas[1],
		e.GasPowerUsed,
		uint64(e.ClaimedTime),
		uint64(e.MedianTime),
	}
	fields32 := []uint32{
		e.Version,
		uint32(e.Epoch),
		uint32(e.Seq),
		uint32(e.Frame),
		uint32(e.Creator),
		uint32(e.Lamport),
		uint32(len(e.Parents)),
	}
	fieldsBool := []bool{
		e.IsRoot,
		isPrevEpochHashEmpty,
		isTxHashEmpty,
	}

	header3 := fast.NewBitArray(
		4, // bits for storing sizes 1-8 of uint64 fields (forced to 4 because fast.BitArray)
		uint(len(fields64)),
	)
	header2 := fast.NewBitArray(
		2, // bits for storing sizes 1-4 of uint32 fields
		uint(len(fields32)+len(fieldsBool)),
	)

	maxBytes := header3.Size() +
		header2.Size() +
		len(fields64)*8 +
		len(fields32)*4 +
		len(e.Parents)*(32-4) + // without idx.Epoch
		common.HashLength + // PrevEpochHash
		common.HashLength + // TxHash
		len(e.Extra)

	raw := make([]byte, maxBytes, maxBytes)

	offset := 0
	header3w := header3.Writer(raw[offset : offset+header3.Size()])
	offset += header3.Size()
	header2w := header2.Writer(raw[offset : offset+header2.Size()])
	offset += header2.Size()
	buf := fast.NewBuffer(raw[offset:])

	for _, i64 := range fields64 {
		n := writeUintCompact(buf, i64, 8)
		header3w.Push(n - 1)
	}
	for _, i32 := range fields32 {
		n := writeUintCompact(buf, uint64(i32), 4)
		header2w.Push(n - 1)
	}
	for _, f := range fieldsBool {
		if f {
			header2w.Push(1)
		} else {
			header2w.Push(0)
		}
	}

	for _, p := range e.Parents {
		buf.Write(p.Bytes()[4:]) // without epoch
	}

	if !isPrevEpochHashEmpty {
		buf.Write(e.PrevEpochHash.Bytes())
	}

	if !isTxHashEmpty {
		buf.Write(e.TxHash.Bytes())
	}

	buf.Write(e.Extra)

	length := header3.Size() + header2.Size() + buf.Position()
	return raw[:length], nil
}

func writeUintCompact(buf *fast.Buffer, v uint64, size int) (bytes int) {
	for i := 0; i < size; i++ {
		buf.WriteByte(byte(v))
		bytes++
		v = v >> 8
		if v == 0 {
			break
		}
	}
	return
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (e *EventHeaderData) UnmarshalBinary(raw []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrInvalidEncoding
		}
	}()

	var (
		isPrevEpochHashEmpty bool
		isTxHashEmpty        bool
	)

	var parentCount uint32

	fields64 := []*uint64{
		&e.GasPowerLeft.Gas[0],
		&e.GasPowerLeft.Gas[1],
		&e.GasPowerUsed,
		(*uint64)(&e.ClaimedTime),
		(*uint64)(&e.MedianTime),
	}
	fields32 := []*uint32{
		&e.Version,
		(*uint32)(&e.Epoch),
		(*uint32)(&e.Seq),
		(*uint32)(&e.Frame),
		(*uint32)(&e.Creator),
		(*uint32)(&e.Lamport),
		&parentCount,
	}
	fieldsBool := []*bool{
		&e.IsRoot,
		&isPrevEpochHashEmpty,
		&isTxHashEmpty,
	}

	header3 := fast.NewBitArray(
		4, // bits for storing sizes 1-8 of uint64 fields (forced to 4 because fast.BitArray)
		uint(len(fields64)),
	)
	header2 := fast.NewBitArray(
		2, // bits for storing sizes 1-4 of uint32 fields
		uint(len(fields32)+len(fieldsBool)),
	)

	offset := 0
	header3r := header3.Reader(raw[offset : offset+header3.Size()])
	offset += header3.Size()
	header2r := header2.Reader(raw[offset : offset+header2.Size()])
	offset += header2.Size()
	raw = raw[offset:]
	buf := fast.NewBuffer(raw)

	for _, i64 := range fields64 {
		n := header3r.Pop() + 1
		*i64, err = readUintCompact(buf, n)
		if err != nil {
			return
		}
	}
	var x uint64
	for _, i32 := range fields32 {
		n := header2r.Pop() + 1
		x, err = readUintCompact(buf, n)
		if err != nil {
			return
		}
		*i32 = uint32(x)
	}
	for _, f := range fieldsBool {
		n := header2r.Pop()
		*f = (n != 0)
	}

	e.Parents = make(hash.Events, parentCount, parentCount)
	for i := uint32(0); i < parentCount; i++ {
		copy(e.Parents[i][:4], e.Epoch.Bytes())
		copy(e.Parents[i][4:], buf.Read(common.HashLength-4)) // without epoch
	}

	if !isPrevEpochHashEmpty {
		e.PrevEpochHash.SetBytes(buf.Read(common.HashLength))
		if e.PrevEpochHash == hash.Zero {
			return ErrNonCanonicalEncoding
		}
	}

	if !isTxHashEmpty {
		e.TxHash.SetBytes(buf.Read(common.HashLength))
		if e.TxHash == EmptyTxHash {
			return ErrNonCanonicalEncoding
		}
	} else {
		e.TxHash = EmptyTxHash
	}

	e.Extra = buf.Read(len(raw) - buf.Position())

	return nil
}

func readUintCompact(buf *fast.Buffer, bytes int) (uint64, error) {
	var (
		v    uint64
		last byte
	)
	for i, b := range buf.Read(bytes) {
		v += uint64(b) << uint(8*i)
		last = b
	}

	if bytes > 1 && last == 0 {
		return 0, ErrNonCanonicalEncoding
	}

	return v, nil
}
