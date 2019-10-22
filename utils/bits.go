package utils

type (
	// BitArray stores only first <bits> of each of <count> numbers.
	BitArray struct {
		bits  uint
		count uint
		vals  []int

		size int
	}

	// BitArrayWriter of numbers to BitArray.
	BitArrayWriter struct {
		BitArray
		offset uint
		raw    []byte

		i   uint
		buf uint16
		n   uint
	}

	// BitArrayReader of numbers from BitArray.
	BitArrayReader struct {
		BitArray
		offset uint
		raw    []byte

		mask uint16
		i    uint
		buf  uint16
		n    uint
	}
)

// NewBitArray makes bits array of int.
func NewBitArray(bits, count uint) *BitArray {
	if bits >= 8 {
		panic("too big size, use bytes")
	}

	return &BitArray{
		bits:  bits,
		count: count,
		vals:  make([]int, count, count),
		size:  calcSize(bits, count),
	}
}

// Writer is a number packer.
func (a *BitArray) Writer(w []byte) *BitArrayWriter {
	if len(w) != a.size {
		panic("need .Size() bytes for writing")
	}

	return &BitArrayWriter{
		BitArray: *a,
		raw:      w,
	}
}

// Reader is a number unpacker.
func (a *BitArray) Reader(r []byte) *BitArrayReader {
	if len(r) != a.size {
		panic("need .Size() bytes for reading")
	}

	return &BitArrayReader{
		BitArray: *a,
		raw:      r,
		mask:     uint16(1<<a.bits) - 1,
	}
}

func calcSize(bits, count uint) int {
	bits = bits * count
	s := bits / 8
	if bits%8 > 0 {
		s++
	}
	return int(s)
}

// Size is a bytes count.
func (a *BitArray) Size() int {
	return a.size
}

// Push bits of number into array.
func (a *BitArrayWriter) Push(v int) {
	if v < 0 {
		panic("only positives accepts")
	}
	if v >= (1 << a.bits) {
		panic("too big number")
	}

	if a.i >= a.count {
		panic("array is full")
	}

	a.buf += uint16(v << a.n)
	a.n += a.bits
	a.raw[a.offset] = byte(a.buf)
	for a.n >= 8 {
		a.offset++
		a.buf = a.buf >> 8
		a.n -= 8
	}

	a.i++
}

// Pop number from array.
func (a *BitArrayReader) Pop() int {
	if a.i >= a.count {
		panic("no numbers more")
	}

	if a.n < a.bits {
		v := a.raw[a.offset]
		a.offset++
		a.buf += uint16(v) << a.n
		a.n += 8
	}
	res := int(a.buf & a.mask)
	a.i++
	a.buf = (a.buf >> a.bits)
	a.n -= a.bits

	return res
}
