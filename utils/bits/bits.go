package bits

type (
	// Array is a bitmap array
	Array struct {
		Bytes []byte
	}

	// Writer of numbers to Array.
	Writer struct {
		*Array
		bitOffset int
	}

	// Reader of numbers from Array.
	Reader struct {
		*Array
		byteOffset int
		bitOffset  int
	}
)

// NewWriter is a bitmap writer
func NewWriter(arr *Array) *Writer {
	return &Writer{
		Array: arr,
	}
}

// NewReader is a bitmap reader
func NewReader(arr *Array) *Reader {
	return &Reader{
		Array: arr,
	}
}

func (a *Writer) byteBitsFree() int {
	return 8 - a.bitOffset
}

func (a *Writer) writeIntoLastByte(v uint) {
	a.Bytes[len(a.Bytes)-1] |= byte(v << a.bitOffset)
}

func zeroTopByteBits(v uint, bits int) uint {
	mask := uint(0xff) >> bits
	return v & mask
}

// Write bits of a number into array.
func (a *Writer) Write(bits int, v uint) {
	if a.bitOffset == 0 {
		a.Bytes = append(a.Bytes, byte(0))
	}
	free := a.byteBitsFree()
	if bits <= free {
		toWrite := bits
		// appending v to the bit array
		a.writeIntoLastByte(v)
		// increment offsets
		if toWrite == free {
			a.bitOffset = 0
		} else {
			a.bitOffset += toWrite
		}
	} else {
		toWrite := free
		clear := a.bitOffset // 8 - free
		// zeroing top `clear` bits and appending result to the bit array
		a.writeIntoLastByte(zeroTopByteBits(v, clear))
		// increment offsets
		a.bitOffset = 0
		a.Write(bits-toWrite, v>>toWrite)
	}
}

func (a *Reader) byteBitsFree() int {
	return 8 - a.bitOffset
}

func (a *Reader) Read(bits int) (v uint) {
	// perform all the checks in the same function to make CPU branch predictor work better
	if bits == 0 {
		return 0
	}
	/*if bits > a.NonReadBits() {
		panic(io.ErrUnexpectedEOF)
	}*/

	free := a.byteBitsFree()
	if bits <= free {
		toRead := bits
		clear := 8 - (a.bitOffset + toRead)
		v = zeroTopByteBits(uint(a.Bytes[a.byteOffset]), clear) >> a.bitOffset
		// increment offsets
		if toRead == free {
			a.bitOffset = 0
			a.byteOffset++
		} else {
			a.bitOffset += toRead
		}
	} else {
		toRead := free
		v = uint(a.Bytes[a.byteOffset]) >> a.bitOffset
		// increment offsets
		a.bitOffset = 0
		a.byteOffset++
		// read rest
		rest := a.Read(bits - toRead)
		v |= rest << toRead
	}
	return
}

func (a *Reader) View(bits int) (v uint) {
	cp := *a
	cpp := &cp
	return cpp.Read(bits)
}

// NonReadBytes returns a number of non-consumed bytes
func (a *Reader) NonReadBytes() int {
	return len(a.Bytes) - a.byteOffset
}

// NonReadBytes returns a number of non-consumed bits
func (a *Reader) NonReadBits() int {
	//return a.nonReadBits
	return a.NonReadBytes()*8 - a.bitOffset
}
