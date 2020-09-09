package bits

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitArrayEmpty(t *testing.T) {
	testBitArray(t, []testWord{}, "empty")
}

func TestBitArrayB0(t *testing.T) {
	testBitArray(t, []testWord{
		{1, 0b0},
	}, "b0")
}

func TestBitArrayB1(t *testing.T) {
	testBitArray(t, []testWord{
		{1, 0b1},
	}, "b1")
}

func TestBitArrayB010101010(t *testing.T) {
	testBitArray(t, []testWord{
		{9, 0b010101010},
	}, "b010101010")
}

func TestBitArrayV01010101010101010(t *testing.T) {
	testBitArray(t, []testWord{
		{17, 0b01010101010101010},
	}, "b01010101010101010")
}

func TestBitArrayRand1(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	for i := 0; i < 50; i++ {
		testBitArray(t, genTestWords(r, 24, 1), fmt.Sprintf("1 bit, case#%d", i))
	}
}

func TestBitArrayRand8(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	for i := 0; i < 50; i++ {
		testBitArray(t, genTestWords(r, 100, 8), fmt.Sprintf("8 bits, case#%d", i))
	}
}

func TestBitArrayRand17(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	for i := 0; i < 50; i++ {
		testBitArray(t, genTestWords(r, 50, 17), fmt.Sprintf("17 bits, case#%d", i))
	}
}

func genTestWords(r *rand.Rand, maxCount int, maxBits int) []testWord {
	count := r.Intn(maxCount)
	words := make([]testWord, count)
	for i := range words {
		if maxBits == 1 {
			words[i].bits = 1
		} else {
			words[i].bits = 1 + r.Intn(maxBits-1)
		}
		words[i].v = uint(r.Intn(1 << words[i].bits))
	}
	return words
}

func bytesToFit(bits int) int {
	if bits%8 == 0 {
		return bits / 8
	}
	return bits/8 + 1
}

type testWord struct {
	bits int
	v    uint
}

func testBitArray(t *testing.T, words []testWord, name string) {
	arr := Array{make([]byte, 0, 100)}
	writer := NewWriter(&arr)
	reader := NewReader(&arr)

	totalBitsWritten := 0
	for _, w := range words {
		writer.Write(w.bits, w.v)
		totalBitsWritten += w.bits
	}
	assert.EqualValuesf(t, bytesToFit(totalBitsWritten), len(arr.Bytes), name)

	totalBitsRead := 0
	for _, w := range words {
		assert.EqualValuesf(t, bytesToFit(totalBitsWritten)*8-totalBitsRead, reader.NonReadBits(), name)
		assert.EqualValuesf(t, bytesToFit(reader.NonReadBits()), reader.NonReadBytes(), name)

		v := reader.Read(w.bits)
		assert.EqualValuesf(t, w.v, v, name)
		totalBitsRead += w.bits

		assert.EqualValuesf(t, bytesToFit(totalBitsWritten)*8-totalBitsRead, reader.NonReadBits(), name)
		assert.EqualValuesf(t, bytesToFit(reader.NonReadBits()), reader.NonReadBytes(), name)
	}

	// read the tail
	assert.Panicsf(t, func() {
		reader.Read(reader.NonReadBits() + 1)
	}, name)
	zero := reader.Read(reader.NonReadBits())
	assert.EqualValuesf(t, uint(0), zero, name)
	assert.EqualValuesf(t, int(0), reader.NonReadBits(), name)
	assert.EqualValuesf(t, int(0), reader.NonReadBytes(), name)
}

func BenchmarkArray_write(b *testing.B) {
	for bits := 1; bits <= 9; bits++ {
		b.Run(fmt.Sprintf("%d bits", bits), func(b *testing.B) {
			b.ResetTimer()

			arr := Array{make([]byte, 0, bytesToFit(bits*b.N))}
			writer := NewWriter(&arr)
			for i := 0; i < b.N; i++ {
				writer.Write(bits, 0xff)
			}
		})
	}
}

func BenchmarkArray_read(b *testing.B) {
	for bits := 1; bits <= 9; bits++ {
		b.Run(fmt.Sprintf("%d bits", bits), func(b *testing.B) {
			b.ResetTimer()

			arr := Array{make([]byte, bytesToFit(bits*b.N))}
			writer := NewReader(&arr)
			for i := 0; i < b.N; i++ {
				_ = writer.Read(bits)
			}
		})
	}
}
