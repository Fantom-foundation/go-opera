package topicsdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPosToBytes(t *testing.T) {
	assertar := assert.New(t)

	for expect := uint8(0); ; /*see break*/ expect += 0x0f {
		bb := posToBytes(expect)
		got := bytesToPos(bb)

		if !assertar.Equal(expect, got) {
			return
		}

		if expect == 0xff {
			break
		}
	}
}

func TestBlocksMask(t *testing.T) {
	assertar := assert.New(t)

	for i, tt := range []struct {
		from uint64
		to   uint64
		mask []byte
	}{
		{0x00110021, 0x00110022, []byte{0, 0, 0, 0, 0x00, 0x11, 0x00}},
		{0x00110022, 0x00110020, []byte{0, 0, 0, 0, 0x00, 0x11, 0x00}},
		{0, 0x0000000000000000, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
		{0, 0xffffffffffffffff, []byte{}},
	} {
		exp := tt.mask
		got := blocksMask(tt.from, tt.to)
		assertar.Equal(exp, got, i)
	}
}
