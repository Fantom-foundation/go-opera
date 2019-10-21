package fast

type BitArray struct {
	buf   			*[]byte
	bitsPartSize   	uint
	offsetPart		uint
}

func NewBitArray(bits int, bytes *[]byte) *BitArray {
	return &BitArray{
		buf:          bytes,
		bitsPartSize: uint(bits),
		offsetPart:   0,
	}
}

func BitArraySizeCalc(bits, count int) int {
	bits = bits * count
	s := bits / 8
	if bits%8 > 0 {
		s++
	}
	return int(s)
}

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

func (a *BitArray) Pop() int {
	byteOffset, bytePartOffset := a.nextPart()
	mask := byte((1 << a.bitsPartSize) - 1)
	return int(((*a.buf)[byteOffset] >> bytePartOffset) & mask)
}

func (a *BitArray) Seek(part int) {
	a.offsetPart = uint(part)
}

func (a *BitArray) nextPart() (byteOffset, bytePartOffset uint) {
	bitsOffset := a.offsetPart * a.bitsPartSize
	byteOffset = bitsOffset/8
	bytePartOffset = bitsOffset%8
	a.offsetPart++

	return
}
