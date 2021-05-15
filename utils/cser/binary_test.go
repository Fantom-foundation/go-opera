package cser

import (
	"errors"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/utils/fast"
)

func TestEmpty(t *testing.T) {
	var (
		buf []byte
		err error
	)

	t.Run("Write", func(t *testing.T) {
		buf, err = MarshalBinaryAdapter(func(w *Writer) error {
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("Read", func(t *testing.T) {
		err = UnmarshalBinaryAdapter(buf, func(r *Reader) error {
			return nil
		})
		require.NoError(t, err)
	})
}

func TestErr(t *testing.T) {
	var (
		buf []byte
	)

	bufCopy := func() []byte {
		bb := make([]byte, len(buf))
		copy(bb, buf)
		return bb
	}

	t.Run("Write", func(t *testing.T) {
		require := require.New(t)

		bb, err := MarshalBinaryAdapter(func(w *Writer) error {
			w.U64(math.MaxUint64)
			return nil
		})
		require.NoError(err)
		buf = append(buf, bb...)

		errExp := errors.New("custom")
		bb, err = MarshalBinaryAdapter(func(w *Writer) error {
			w.Bool(false)
			return errExp
		})
		require.Equal(errExp, err)
		buf = append(buf, bb...)
	})

	t.Run("Read nil", func(t *testing.T) {
		require := require.New(t)
		// nothing unmarshal
		err := UnmarshalBinaryAdapter(nil, func(r *Reader) error {
			return nil
		})
		require.Equal(ErrMalformedEncoding, err)
	})

	t.Run("Read err", func(t *testing.T) {
		require := require.New(t)

		errExp := errors.New("custom")
		// unmarshal
		err := UnmarshalBinaryAdapter(buf, func(r *Reader) error {
			require.Equal(uint64(math.MaxUint64), r.U64())
			return errExp
		})
		require.Equal(errExp, err)
	})

	t.Run("Read 0", func(t *testing.T) {
		require := require.New(t)
		// unpack
		_, bbytes, err := binaryToCSER(bufCopy())
		require.NoError(err)
		// pack with wrong bits size
		corrupted := fast.NewWriter(bbytes)
		sizeWriter := fast.NewWriter(make([]byte, 0, 4))
		writeUint64Compact(sizeWriter, uint64(len(bbytes)+1))
		corrupted.Write(reversed(sizeWriter.Bytes()))
		// corrupted unpack
		_, _, err = binaryToCSER(corrupted.Bytes())
		require.Equal(ErrMalformedEncoding, err)
		// corrupted unmarshal
		err = UnmarshalBinaryAdapter(corrupted.Bytes(), func(r *Reader) error {
			require.Equal(uint64(math.MaxUint64), r.U64())
			return nil
		})
		require.Equal(ErrMalformedEncoding, err)
	})

	repackWithDefect := func(
		defect func(bbits, bbytes *[]byte) (expected error),
	) func(t *testing.T) {
		return func(t *testing.T) {
			require := require.New(t)
			// unpack
			bbits, bbytes, err := binaryToCSER(bufCopy())
			require.NoError(err)
			// pack with defect
			errExp := defect(&bbits.Bytes, &bbytes)
			corrupted, err := binaryFromCSER(bbits, bbytes)
			require.NoError(err)
			// corrupted unmarshal
			err = UnmarshalBinaryAdapter(corrupted, func(r *Reader) error {
				_ = r.U64()
				return nil
			})
			require.Equal(errExp, err)
		}
	}

	t.Run("Read 1", repackWithDefect(func(bbits, bbytes *[]byte) (expected error) {
		// no defect
		return nil
	}))

	t.Run("Read 2", repackWithDefect(func(bbits, bbytes *[]byte) (expected error) {
		*bbytes = append(*bbytes, 0xFF)
		return ErrNonCanonicalEncoding
	}))

	t.Run("Read 3", repackWithDefect(func(bbits, bbytes *[]byte) (expected error) {
		*bbits = append(*bbits, 0x0F)
		return ErrNonCanonicalEncoding
	}))

	t.Run("Read 4", repackWithDefect(func(bbits, bbytes *[]byte) (expected error) {
		*bbytes = (*bbytes)[:len(*bbytes)-1]
		return ErrNonCanonicalEncoding
	}))
}

func TestVals(t *testing.T) {
	var (
		buf []byte
		err error
	)
	var (
		expBigInt     = []*big.Int{big.NewInt(0), big.NewInt(0xFFFFF)}
		expBool       = []bool{true, false}
		expFixedBytes = [][]byte{[]byte{}, randBytes(0xFF)}
		expSliceBytes = [][]byte{[]byte{}, randBytes(0xFF)}
		expU8         = []uint8{0, 1, 0xFF}
		expU16        = []uint16{0, 1, 0xFFFF}
		expU32        = []uint32{0, 1, 0xFFFFFFFF}
		expU64        = []uint64{0, 1, 0xFFFFFFFFFFFFFFFF}
		expVarUint    = []uint64{0, 1, 0xFFFFFFFFFFFFFFFF}
		expI64        = []int64{0, 1, math.MinInt64, math.MaxInt64}
		expU56        = []uint64{0, 1, 1<<(8*7) - 1}
	)

	t.Run("Write", func(t *testing.T) {
		require := require.New(t)

		buf, err = MarshalBinaryAdapter(func(w *Writer) error {
			for _, v := range expBigInt {
				w.BigInt(v)
			}
			for _, v := range expBool {
				w.Bool(v)
			}
			for _, v := range expFixedBytes {
				w.FixedBytes(v)
			}
			for _, v := range expSliceBytes {
				w.SliceBytes(v)
			}
			for _, v := range expU8 {
				w.U8(v)
			}
			for _, v := range expU16 {
				w.U16(v)
			}
			for _, v := range expU32 {
				w.U32(v)
			}
			for _, v := range expU64 {
				w.U64(v)
			}
			for _, v := range expVarUint {
				w.VarUint(v)
			}
			for _, v := range expI64 {
				w.I64(v)
			}
			for _, v := range expU56 {
				w.U56(v)
			}
			return nil
		})
		require.NoError(err)
	})

	t.Run("Read", func(t *testing.T) {
		require := require.New(t)

		err = UnmarshalBinaryAdapter(buf, func(r *Reader) error {
			for i, exp := range expBigInt {
				got := r.BigInt()
				require.Equal(exp, got, i)
			}
			for i, exp := range expBool {
				got := r.Bool()
				require.Equal(exp, got, i)
			}
			for i, exp := range expFixedBytes {
				got := make([]byte, len(exp))
				r.FixedBytes(got)
				require.Equal(exp, got, i)
			}
			for i, exp := range expSliceBytes {
				got := r.SliceBytes()
				require.Equal(exp, got, i)
			}
			for i, exp := range expU8 {
				got := r.U8()
				require.Equal(exp, got, i)
			}
			for i, exp := range expU16 {
				got := r.U16()
				require.Equal(exp, got, i)
			}
			for i, exp := range expU32 {
				got := r.U32()
				require.Equal(exp, got, i)
			}
			for i, exp := range expU64 {
				got := r.U64()
				require.Equal(exp, got, i)
			}
			for i, exp := range expVarUint {
				got := r.VarUint()
				require.Equal(exp, got, i)
			}
			for i, exp := range expI64 {
				got := r.I64()
				require.Equal(exp, got, i)
			}
			for i, exp := range expU56 {
				got := r.U56()
				require.Equal(exp, got, i)
			}
			return nil
		})
		require.NoError(err)
	})
}

func TestBadVals(t *testing.T) {
	var (
		buf []byte
		err error
	)
	var (
		expBigInt     = []*big.Int{nil}
		expFixedBytes = [][]byte{nil}
		expSliceBytes = [][]byte{nil}
		expU56        = []uint64{1 << (8 * 7), math.MaxUint64}
	)

	t.Run("Write", func(t *testing.T) {
		require := require.New(t)

		buf, err = MarshalBinaryAdapter(func(w *Writer) error {
			for _, v := range expBigInt {
				require.Panics(func() {
					w.BigInt(v)
				})
			}
			for _, v := range expFixedBytes {
				w.FixedBytes(v)
			}
			for _, v := range expSliceBytes {
				w.SliceBytes(v)
			}
			for _, v := range expU56 {
				require.Panics(func() {
					w.U56(v)
				})
			}
			return nil
		})
		require.NoError(err)
	})

	t.Run("Read", func(t *testing.T) {
		require := require.New(t)

		err = UnmarshalBinaryAdapter(buf, func(r *Reader) error {
			for _, _ = range expBigInt {
				// skip
			}
			for i, exp := range expFixedBytes {
				got := make([]byte, len(exp))
				r.FixedBytes(got)
				require.NotEqual(exp, got, i)
				require.Equal(len(exp), len(got), i)
			}
			for i, exp := range expSliceBytes {
				got := r.SliceBytes()
				require.NotEqual(exp, got, i)
				require.Equal(len(exp), len(got), i)
			}
			for _, _ = range expU56 {
				// skip
			}
			return nil
		})
		require.NoError(err)
	})
}

func randBytes(n int) []byte {
	bb := make([]byte, n)
	_, err := rand.Read(bb)
	if err != nil {
		panic(err)
	}
	return bb
}
