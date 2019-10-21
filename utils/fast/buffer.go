package fast

import "errors"

type Buffer struct {
	buf *[]byte
	offset     	int
}

func NewBuffer(buf *[]byte) *Buffer {
	b := Buffer{
		buf:        buf,
		offset:     0,
	}

	return &b
}

func (b *Buffer) Write(src []byte) {
	size := len(src)
	b.WriteLen(src, size)
}

func (b *Buffer) WriteByte(src byte) {
	(*b.buf)[b.offset] = src
	b.offset++
}

func (b *Buffer) WriteLen(src []byte, size int) {
	copy((*b.buf)[b.offset:b.offset+ size], src)
	b.offset += size
}

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

func (b *Buffer) ReadByte() byte {
	res := (*b.buf)[b.offset]
	b.offset++
	return res
}

func (b *Buffer) Seek(off int) error {
	if off < 0 || off > len(*b.buf) {
		return errors.New("Index out of range")
	}
	b.offset = off

	return nil
}

func (b *Buffer) Bytes() []byte {
	return (*b.buf)[0:b.offset]
}

func (b *Buffer) BytesLen() int {
	return b.offset
}
