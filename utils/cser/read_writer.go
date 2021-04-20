package cser

import (
	"errors"
	"math/big"

	"github.com/Fantom-foundation/go-opera/utils/bits"
	"github.com/Fantom-foundation/go-opera/utils/fast"
)

var (
	ErrNonCanonicalEncoding = errors.New("non canonical encoding")
	ErrMalformedEncoding    = errors.New("malformed encoding")
)

type Writer struct {
	BitsW  *bits.Writer
	BytesW *fast.Writer
}

type Reader struct {
	BitsR  *bits.Reader
	BytesR *fast.Reader
}

func NewWriter() *Writer {
	bbits := &bits.Array{Bytes: make([]byte, 0, 32)}
	bbytes := make([]byte, 0, 200)
	return &Writer{
		BitsW:  bits.NewWriter(bbits),
		BytesW: fast.NewWriter(bbytes),
	}
}

func writeUint64Compact(bytesW *fast.Writer, v uint64) {
	for i := 0; ; i++ {
		chunk := v & 0b01111111
		v = v >> 7
		if v == 0 {
			// stop flag
			chunk |= 0b10000000
		}
		bytesW.WriteByte(byte(chunk))
		if v == 0 {
			break
		}
	}
	return
}

func readUint64Compact(bytesR *fast.Reader) uint64 {
	v := uint64(0)
	stop := false
	for i := 0; !stop; i++ {
		chunk := uint64(bytesR.ReadByte())
		stop = (chunk & 0b10000000) != 0
		word := chunk & 0b01111111
		v |= word << (i * 7)
		// last byte cannot be zero
		if i > 0 && stop && word == 0 {
			panic(ErrNonCanonicalEncoding)
		}
	}

	return v
}

func writeUint64BitCompact(bytesW *fast.Writer, v uint64, minSize int) (size int) {
	for size < minSize || v != 0 {
		bytesW.WriteByte(byte(v))
		size++
		v = v >> 8
	}
	return
}

func readUint64BitCompact(bytesR *fast.Reader, size int) uint64 {
	var (
		v    uint64
		last byte
	)
	buf := bytesR.Read(size)
	for i, b := range buf {
		v |= uint64(b) << uint(8*i)
		last = b
	}

	if size > 1 && last == 0 {
		panic(ErrNonCanonicalEncoding)
	}

	return v
}

func (r *Reader) U8() uint8 {
	return r.BytesR.ReadByte()
}

func (w *Writer) U8(v uint8) {
	w.BytesW.WriteByte(v)
}

func (r *Reader) readU64_bits(minSize int, bitsForSize int) uint64 {
	size := r.BitsR.Read(bitsForSize)
	size += uint(minSize)
	return readUint64BitCompact(r.BytesR, int(size))
}

func (w *Writer) writeU64_bits(minSize int, bitsForSize int, v uint64) {
	size := writeUint64BitCompact(w.BytesW, v, minSize)
	w.BitsW.Write(bitsForSize, uint(size-minSize))
}

func (r *Reader) U16() uint16 {
	v64 := r.readU64_bits(1, 1)
	return uint16(v64)
}

func (w *Writer) U16(v uint16) {
	w.writeU64_bits(1, 1, uint64(v))
}

func (r *Reader) U32() uint32 {
	v64 := r.readU64_bits(1, 2)
	return uint32(v64)
}

func (w *Writer) U32(v uint32) {
	w.writeU64_bits(1, 2, uint64(v))
}

func (r *Reader) U64() uint64 {
	return r.readU64_bits(1, 3)
}

func (w *Writer) U64(v uint64) {
	w.writeU64_bits(1, 3, v)
}

func (r *Reader) VarUint() uint64 {
	return r.readU64_bits(1, 3)
}

func (w *Writer) VarUint(v uint64) {
	w.writeU64_bits(1, 3, v)
}

func (r *Reader) I64() int64 {
	neg := r.Bool()
	abs := r.U64()
	if neg && abs == 0 {
		panic(ErrNonCanonicalEncoding)
	}
	if neg {
		return -int64(abs)
	}
	return int64(abs)
}

func (w *Writer) I64(v int64) {
	w.Bool(v < 0)
	if v < 0 {
		w.U64(uint64(-v))
	} else {
		w.U64(uint64(v))
	}
}

func (r *Reader) U56() uint64 {
	return r.readU64_bits(0, 3)
}

func (w *Writer) U56(v uint64) {
	const max = 1<<(8*7) - 1
	if v > max {
		panic("Value too big")
	}
	w.writeU64_bits(0, 3, v)
}

func (r *Reader) Bool() bool {
	u8 := r.BitsR.Read(1)
	return u8 != 0
}

func (w *Writer) Bool(v bool) {
	u8 := uint(0)
	if v {
		u8 = 1
	}
	w.BitsW.Write(1, u8)
}

func (r *Reader) FixedBytes(v []byte) {
	buf := r.BytesR.Read(len(v))
	copy(v, buf)
}

func (w *Writer) FixedBytes(v []byte) {
	w.BytesW.Write(v)
}

func (r *Reader) SliceBytes() []byte {
	// read slice size
	size := r.U56()
	buf := make([]byte, size)
	// read slice content
	r.FixedBytes(buf)
	return buf
}

func (w *Writer) SliceBytes(v []byte) {
	// write slice size
	w.U56(uint64(len(v)))
	// write slice content
	w.FixedBytes(v)
}

// PaddedBytes returns a slice with length of the slice is at least n bytes.
func PaddedBytes(b []byte, n int) []byte {
	if len(b) >= n {
		return b
	}
	padding := make([]byte, n-len(b))
	return append(padding, b...)
}

func (w *Writer) BigInt(v *big.Int) {
	// serialize as an ordinary slice
	bigBytes := []byte{}
	if v.Sign() != 0 {
		bigBytes = v.Bytes()
	}
	w.SliceBytes(bigBytes)
}

func (r *Reader) BigInt() *big.Int {
	// deserialize as an ordinary slice
	buf := r.SliceBytes()
	if len(buf) == 0 {
		return new(big.Int)
	}
	return new(big.Int).SetBytes(buf)
}
