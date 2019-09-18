package poset

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestFrameInfoSerialization(t *testing.T) {
	assertar := assert.New(t)

	f0 := &FrameInfo{
		TimeOffset: math.MaxInt64 / 2,
		TimeRatio:  math.MaxUint64,
	}
	buf, err := rlp.EncodeToBytes(f0)
	assertar.NoError(err)

	f1 := &FrameInfo{}
	err = rlp.DecodeBytes(buf, f1)
	assertar.NoError(err)

	assertar.EqualValues(f0, f1)
}

func TestFrameInfoSerializationSigned(t *testing.T) {
	assertar := assert.New(t)

	f0 := &FrameInfo{
		TimeOffset: -math.MaxInt64 / 2,
		TimeRatio:  0,
	}
	buf, err := rlp.EncodeToBytes(f0)
	assertar.NoError(err)

	f1 := &FrameInfo{}
	err = rlp.DecodeBytes(buf, f1)
	assertar.NoError(err)

	assertar.EqualValues(f0, f1)
}
