package fast

// Buffer is a fixed-sized buffer of byte with Read and Write methods.
type Buffer struct {
	buf    *[]byte
	offset int
}

// NewBuffer wraps bytes with buffer.
func NewBuffer(bb []byte) *Buffer {
	return &Buffer{
		buf:    &bb,
		offset: 0,
	}
}

// WriteByte to the buffer.
func (b *Buffer) WriteByte(v byte) {
	(*b.buf)[b.offset] = v
	b.offset++
}

// Write the byte to the buffer.
func (b *Buffer) Write(v []byte) {
	n := copy((*b.buf)[b.offset:], v)
	b.offset += n

	if n != len(v) {
		panic("buffer overflow")
	}
}

// Read n bytes.
func (b *Buffer) Read(n int) []byte {
	var res []byte
	res = (*b.buf)[b.offset : b.offset+n]
	b.offset += n

	return res
}

// Position of internal cursor.
func (b *Buffer) Position() int {
	return b.offset
}
