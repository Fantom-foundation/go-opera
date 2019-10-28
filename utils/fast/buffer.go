package fast

import "errors"

// Buffer is small fast byte buffer realization
type Buffer struct {
	buf *[]byte
	offset     	int
}

// NewBuffer create new fast.Buffer
func NewBuffer(buf *[]byte) *Buffer {
	b := Buffer{
		buf:        buf,
		offset:     0,
	}

	return &b
}

// Write method write byte slice in to buffer
func (b *Buffer) Write(src []byte) {
	size := len(src)
	b.WriteLen(src, size)
}

// WriteByte method write one byte in to buffer
func (b *Buffer) WriteByte(src byte) error {
	(*b.buf)[b.offset] = src
	b.offset++
	return nil
}

// WriteLen method write byte sequence in to buffer with fixed size
func (b *Buffer) WriteLen(src []byte, size int) {
	copy((*b.buf)[b.offset:b.offset+ size], src)
	b.offset += size
}

// Read method read byte slice with fixed len from buffer
func (b *Buffer) Read(size int) (result []byte) {
	if size <= 0 {
		return []byte{}
	}
	if b.offset + size > len(*b.buf) {
		size = len(*b.buf) - b.offset
	}
	result = (*b.buf)[b.offset:b.offset+size]
	b.offset += size
	return
}

// ReadByte method read one byte from buffer
func (b *Buffer) ReadByte() (byte, error) {
	res := (*b.buf)[b.offset]
	b.offset++
	return res, nil
}

// Seek move internal offset to position in parameter
func (b *Buffer) Seek(off int) error {
	if off < 0 || off > len(*b.buf) {
		return errors.New("Index out of range")
	}
	b.offset = off

	return nil
}

// Bytes return byte slice from start of buffer to current offset
func (b *Buffer) Bytes() []byte {
	return (*b.buf)[0:b.offset]
}

// BytesLen return current offset position (len of byte slice returned from Bytes() method)
func (b *Buffer) BytesLen() int {
	return b.offset
}
