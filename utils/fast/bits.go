package fast

// BitArray used for save int in bits sets with fixed sizes: 1, 2, 4 bits
type BitArray struct {
	buf          *[]byte
	bitsPartSize uint
	offsetPart   uint
}

// NewBitArray create new bit array for parts size (in bits) in exists byte slice
func NewBitArray(bits int, bytes *[]byte) *BitArray {
	return &BitArray{
		buf:          bytes,
		bitsPartSize: uint(bits),
		offsetPart:   0,
	}
}

// BitArraySizeCalc calculate bytes count of buffer required for save 'count' int values in BitArray with 'bits' part size
func BitArraySizeCalc(bits, count int) int {
	bits = bits * count
	s := bits / 8
	if bits%8 > 0 {
		s++
	}
	return int(s)
}

// Push method push int value in buffer with pack it to BitArray packed size bits
func (a *BitArray) Push(v int) {
	if v < 0 {
		panic("positives only")
	}
	if v >= (1 << a.bitsPartSize) {
		panic("too big value")
	}

	byteOffset, bytePartOffset := a.nextPart()

	(*a.buf)[byteOffset] |= byte(v << bytePartOffset)
}

// Pop method pop int value from buffer with unpack it from BitArray packed size bits
func (a *BitArray) Pop() int {
	byteOffset, bytePartOffset := a.nextPart()
	mask := byte((1 << a.bitsPartSize) - 1)
	return int(((*a.buf)[byteOffset] >> bytePartOffset) & mask)
}

// Seek move current offset of part to 'part' value
func (a *BitArray) Seek(part int) {
	a.offsetPart = uint(part)
}

// nextPart return offset in bytes in buffer and number part inside byte for next bits set for value
func (a *BitArray) nextPart() (byteOffset, bytePartOffset uint) {
	bitsOffset := a.offsetPart * a.bitsPartSize
	byteOffset = bitsOffset / 8
	bytePartOffset = bitsOffset % 8
	a.offsetPart++

	return
}
