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
