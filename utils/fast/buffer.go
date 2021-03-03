package fast

type Reader struct {
	buf    []byte
	offset int
}

type Writer struct {
	buf []byte
}

// NewReader wraps bytes with reading buffer.
func NewReader(bb []byte) *Reader {
	return &Reader{
		buf:    bb,
		offset: 0,
	}
}

// NewWriter wraps bytes with writing buffer.
func NewWriter(bb []byte) *Writer {
	return &Writer{
		buf: bb,
	}
}

// WriteByte to the buffer.
func (b *Writer) WriteByte(v byte) {
	b.buf = append(b.buf, v)
}

// Write the byte to the buffer.
func (b *Writer) Write(v []byte) {
	b.buf = append(b.buf, v...)
}

// Read n bytes.
func (b *Reader) Read(n int) []byte {
	var res []byte
	res = b.buf[b.offset : b.offset+n]
	b.offset += n

	return res
}

// ReadByte reads 1 byte.
func (b *Reader) ReadByte() byte {
	var res byte
	res = b.buf[b.offset]
	b.offset++

	return res
}

// Position of internal cursor.
func (b *Reader) Position() int {
	return b.offset
}

// Bytes of internal buffer
func (b *Reader) Bytes() []byte {
	return b.buf
}

// Bytes of internal buffer
func (b *Writer) Bytes() []byte {
	return b.buf
}

// Empty returns true if the whole buffer is consumed
func (b *Reader) Empty() bool {
	return len(b.buf) == b.offset
}
