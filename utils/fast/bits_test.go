package fast

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitArray(t *testing.T) {
	for _, i := range []uint{1, 2, 4} {
		testBitArray(t, i)
	}
}

func testBitArray(t *testing.T, bits uint) {

	expect := rand.Perm(1 << bits)
	count := len(expect)
	got := make([]int, count, count)

	arr := NewBitArray(bits, uint(count))
	raw := make([]byte, arr.Size(), arr.Size())

	t.Logf("Bits: %d, len(raw) = %d", bits, len(raw))

	writer := arr.Writer(&raw)
	for _, v := range expect {
		writer.Push(v)
	}

	t.Logf("raw: %v", raw)

	reader := arr.Reader(&raw)
	for i := 0; i < count; i++ {
		got[i] = reader.Pop()
	}

	assert.EqualValuesf(t, expect, got, "bits count: %d", bits)

	assert.EqualValues(t, len(raw), arr.Size(), ".Size()")
}
