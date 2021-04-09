package cser

import (
	"github.com/Fantom-foundation/go-opera/utils/bits"
	"github.com/Fantom-foundation/go-opera/utils/fast"
)

func MarshalBinaryAdapter(marshalCser func(writer *Writer) error) ([]byte, error) {
	bodyBits := &bits.Array{Bytes: make([]byte, 0, 32)}
	bodyBytes := fast.NewWriter(make([]byte, 0, 200))
	bodyWriter := &Writer{
		BitsW:  bits.NewWriter(bodyBits),
		BytesW: bodyBytes,
	}
	err := marshalCser(bodyWriter)
	if err != nil {
		return nil, err
	}

	bodyBytes.Write(bodyBits.Bytes)
	// write bits size
	sizeWriter := fast.NewWriter(make([]byte, 0, 4))
	writeUint64Compact(sizeWriter, uint64(len(bodyBits.Bytes)))
	bodyBytes.Write(reversed(sizeWriter.Bytes()))

	return bodyBytes.Bytes(), nil
}

func binaryToCSER(raw []byte) (bodyBits *bits.Array, bodyBytes []byte, err error) {
	// read bitsArray size
	bitsSizeBuf := reversed(tail(raw, 9))
	bitsSizeReader := fast.NewReader(bitsSizeBuf)
	bitsSize := readUint64Compact(bitsSizeReader)
	raw = raw[:len(raw)-bitsSizeReader.Position()]

	if uint64(len(raw)) < bitsSize {
		return nil, nil, ErrMalformedEncoding
	}

	bodyBits = &bits.Array{Bytes: raw[uint64(len(raw))-bitsSize:]}
	bodyBytes = raw[:uint64(len(raw))-bitsSize]
	return bodyBits, bodyBytes, nil
}

func UnmarshalBinaryAdapter(raw []byte, unmarshalCser func(reader *Reader) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrMalformedEncoding
		}
	}()
	bodyBits, bodyBytes_, err := binaryToCSER(raw)
	if err != nil {
		return err
	}
	bodyBytes := fast.NewReader(bodyBytes_)

	bodyReader := &Reader{
		BitsR:  bits.NewReader(bodyBits),
		BytesR: bodyBytes,
	}
	err = unmarshalCser(bodyReader)
	if err != nil {
		return err
	}

	// check that everything is read
	if bodyReader.BitsR.NonReadBytes() > 1 {
		return ErrNonCanonicalEncoding
	}
	tail := bodyReader.BitsR.Read(bodyReader.BitsR.NonReadBits())
	if tail != 0 {
		return ErrNonCanonicalEncoding
	}
	if !bodyReader.BytesR.Empty() {
		return ErrNonCanonicalEncoding
	}

	return nil
}

func tail(b []byte, cap int) []byte {
	if len(b) > cap {
		return b[len(b)-cap:]
	}
	return b
}

func reversed(b []byte) []byte {
	reversed := make([]byte, len(b))
	for i, v := range b {
		reversed[len(b)-1-i] = v
	}
	return reversed
}
