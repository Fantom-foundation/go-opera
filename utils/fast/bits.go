package fast

type (
	// BitArray stores only first <1,2,4> bits of each of <count> numbers.
	BitArray struct {
		bits  uint
		count uint

		offset uint
		size   int
	}

	// BitArrayWriter of numbers to BitArray.
	BitArrayWriter struct {
		BitArray
		raw *[]byte
	}

	// BitArrayReader of numbers from BitArray.
	BitArrayReader struct {
		BitArray
		raw *[]byte
	}
)

// NewBitArray makes bits array of int.
func NewBitArray(bits, count uint) *BitArray {
	if bits != 1 && bits != 2 && bits != 4 {
		panic("use 1 or 2 or 4 bits")
	}

	return &BitArray{
		bits:  bits,
		count: count,

		size: calcSize(bits, count),
	}
}

// Writer is a number packer.
func (a *BitArray) Writer(w []byte) *BitArrayWriter {
	if len(w) != a.size {
		panic("need .Size() bytes for writing")
	}

	return &BitArrayWriter{
		BitArray: *a,
		raw:      &w,
	}
}

// Reader is a number unpacker.
func (a *BitArray) Reader(r []byte) *BitArrayReader {
	if len(r) != a.size {
		panic("need .Size() bytes for reading")
	}

	return &BitArrayReader{
		BitArray: *a,
		raw:      &r,
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

	byteOffset, bytePartOffset := a.nextPart()
	(*a.raw)[byteOffset] |= byte(v << bytePartOffset)
}

// Pop number from array.
func (a *BitArrayReader) Pop() int {
	byteOffset, bytePartOffset := a.nextPart()
	mask := byte((1 << a.bits) - 1)
	return int(((*a.raw)[byteOffset] >> bytePartOffset) & mask)
}

// nextPart return buffer offset in bytes and number of next byte-part.
func (a *BitArray) nextPart() (byteOffset, bytePartOffset uint) {
	bitsOffset := a.offset * a.bits
	byteOffset = bitsOffset / 8
	bytePartOffset = bitsOffset % 8
	a.offset++

	return
}
