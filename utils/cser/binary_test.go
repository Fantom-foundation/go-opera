package cser

import (
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmpty(t *testing.T) {
	var (
		raw []byte
		err error
	)

	t.Run("Write", func(t *testing.T) {
		raw, err = MarshalBinaryAdapter(func(w *Writer) error {
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("Read", func(t *testing.T) {
		err = UnmarshalBinaryAdapter(raw, func(r *Reader) error {
			return nil
		})
		require.NoError(t, err)
	})
}

func TestVals(t *testing.T) {
	var (
		raw []byte
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

		raw, err = MarshalBinaryAdapter(func(w *Writer) error {
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

		err = UnmarshalBinaryAdapter(raw, func(r *Reader) error {
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
		raw []byte
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

		raw, err = MarshalBinaryAdapter(func(w *Writer) error {
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

		err = UnmarshalBinaryAdapter(raw, func(r *Reader) error {
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
