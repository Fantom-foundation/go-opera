package cser

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/utils/bits"
	"github.com/Fantom-foundation/go-opera/utils/fast"
)

func TestUint64Compact(t *testing.T) {
	require := require.New(t)
	var (
		canonical    = []byte{0b01111111, 0b11111111}
		nonCanonical = []byte{0b01111111, 0b01111111, 0b10000000}
	)

	r := fast.NewReader(canonical)
	got := readUint64Compact(r)
	require.Equal(uint64(0x3fff), got)

	r = fast.NewReader(nonCanonical)
	require.Panics(func() {
		_ = readUint64Compact(r)
	})
}

func TestUint64BitCompact(t *testing.T) {
	require := require.New(t)
	var (
		canonical    = []byte{0b11111111, 0b00111111}
		nonCanonical = []byte{0b01111111, 0b01111111, 0b00000000}
	)

	r := fast.NewReader(canonical)
	got := readUint64BitCompact(r, len(canonical))
	require.Equal(uint64(0x3fff), got)

	r = fast.NewReader(nonCanonical)
	require.Panics(func() {
		_ = readUint64BitCompact(r, len(nonCanonical))
	})
}

func TestI64(t *testing.T) {
	require := require.New(t)

	w := NewWriter()

	canonical := w.I64
	nonCanonical := func(v int64) {
		w.Bool(v <= 0)
		if v < 0 {
			w.U64(uint64(-v))
		} else {
			w.U64(uint64(v))
		}
	}

	canonical(0)
	nonCanonical(0)

	r := &Reader{
		BitsR:  bits.NewReader(w.BitsW.Array),
		BytesR: fast.NewReader(w.BytesW.Bytes()),
	}

	got := r.I64()
	require.Zero(got)

	require.Panics(func() {
		_ = r.I64()
	})
}

func TestPaddedBytes(t *testing.T) {
	require := require.New(t)

	var data = []struct {
		In  []byte
		N   int
		Exp []byte
	}{
		{In: nil, N: 0, Exp: nil},
		{In: []byte{}, N: 0, Exp: []byte{}},
		{In: nil, N: 1, Exp: []byte{0}},
		{In: []byte{}, N: 1, Exp: []byte{0}},
		{In: []byte{10, 20}, N: 1, Exp: []byte{10, 20}},
		{In: []byte{10, 20}, N: 4, Exp: []byte{0, 0, 10, 20}},
	}

	for i, d := range data {
		got := PaddedBytes(d.In, d.N)
		require.Equal(d.Exp, got, i)
	}
}
