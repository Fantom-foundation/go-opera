package fast

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestBitArray(t *testing.T) {
	for _, i := range []int{2, 4, 8} {
		testBitArray(t, i)
	}
}

func testBitArray(t *testing.T, bits int) {
	expect := rand.Perm(1 << uint(bits))
	count := len(expect)

	size := BitArraySizeCalc(bits, count)

	buf := make([]byte, size)

	arr := NewBitArray(bits, &buf)
	for _, v := range expect {
		arr.Push(v)
	}

	arr.Seek(0)
	for _, v := range expect {
		got := arr.Pop()
		assert.EqualValuesf(t, v, got, "bits count: %d", bits)
	}
}
