package topicsdb

import (
	"testing"

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
