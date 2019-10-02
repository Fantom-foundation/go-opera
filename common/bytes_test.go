package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/common/littleendian"
)

func Test_IntToBytes(t *testing.T) {
	assertar := assert.New(t)

	for _, n1 := range []uint64{
		0,
		9,
		0xF000000000000000,
		0x000000000000000F,
		0xFFFFFFFFFFFFFFFF,
		47528346792,
	} {
		{
			b := bigendian.Int64ToBytes(n1)
			n2 := bigendian.BytesToInt64(b)
			assertar.Equal(n1, n2)
		}
		{
			b := littleendian.Int64ToBytes(n1)
			n2 := littleendian.BytesToInt64(b)
			assertar.Equal(n1, n2)
		}
	}
	for _, n1 := range []uint32{
		0,
		9,
		0xFFFFFFFF,
		475283467,
	} {
		{
			b := bigendian.Int32ToBytes(n1)
			n2 := bigendian.BytesToInt32(b)
			assertar.Equal(n1, n2)
		}
		{
			b := littleendian.Int32ToBytes(n1)
			n2 := littleendian.BytesToInt32(b)
			assertar.Equal(n1, n2)
		}
	}
}
