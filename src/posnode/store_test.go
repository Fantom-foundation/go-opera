package posnode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IntToBytes(t *testing.T) {
	assertar := assert.New(t)

	for _, n1 := range []uint64{
		0,
		9,
		0xFFFFFFFFFFFFFF,
		47528346792,
	} {
		b := intToBytes(n1)
		n2 := bytesToInt(b)
		assertar.Equal(n1, n2)
	}
}
