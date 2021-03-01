package topicsdb

import (
	"testing"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/stretchr/testify/require"
)

func TestPosToBytes(t *testing.T) {
	require := require.New(t)

	for i := 0xff / 0x0f; i >= 0; i-- {
		expect := uint8(0x0f * i)
		bb := posToBytes(expect)
		got := bytesToPos(bb)

		require.Equal(expect, got)
	}
}

func TestBlocksMask(t *testing.T) {
	require := require.New(t)

	for i, tt := range []struct {
		from idx.Block
		to   idx.Block
		mask []byte
	}{
		{0x00110021, 0x00110022, []byte{0, 0, 0, 0, 0x00, 0x11, 0x00}},
		{0x00110022, 0x00110020, []byte{0, 0, 0, 0, 0x00, 0x11, 0x00}},
		{0, 0x0000000000000000, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
		{0, 0xffffffffffffffff, []byte{}},
	} {
		exp := tt.mask
		got := blocksMask(tt.from, tt.to)
		require.Equal(exp, got, i)
	}
}
