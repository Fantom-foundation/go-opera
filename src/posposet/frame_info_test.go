package posposet

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFrameInfoSerialization(t *testing.T) {
	assertar := assert.New(t)

	f0 := &FrameInfo{
		TimeOffset: 3,
		TimeRatio:  1,
	}
	buf, err := rlp.EncodeToBytes(f0)
	assertar.NoError(err)

	f1 := &FrameInfo{}
	err = rlp.DecodeBytes(buf, f1)
	assertar.NoError(err)

	assertar.EqualValues(f0, f1)
}
