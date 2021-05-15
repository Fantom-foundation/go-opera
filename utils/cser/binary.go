package cser

import (
	"github.com/Fantom-foundation/go-opera/utils/bits"
	"github.com/Fantom-foundation/go-opera/utils/fast"
)

func MarshalBinaryAdapter(marshalCser func(*Writer) error) ([]byte, error) {
	w := NewWriter()
	err := marshalCser(w)
	if err != nil {
		return nil, err
	}

	return binaryFromCSER(w.BitsW.Array, w.BytesW.Bytes())
}

// binaryFromCSER packs body bytes and bits into raw
func binaryFromCSER(bbits *bits.Array, bbytes []byte) (raw []byte, err error) {
	bodyBytes := fast.NewWriter(bbytes)
	bodyBytes.Write(bbits.Bytes)
	// write bits size
	sizeWriter := fast.NewWriter(make([]byte, 0, 4))
	writeUint64Compact(sizeWriter, uint64(len(bbits.Bytes)))
	bodyBytes.Write(reversed(sizeWriter.Bytes()))
	return bodyBytes.Bytes(), nil
}

// binaryToCSER unpacks raw on body bytes and bits
func binaryToCSER(raw []byte) (bbits *bits.Array, bbytes []byte, err error) {
	// read bitsArray size
	bitsSizeBuf := reversed(tail(raw, 9))
	bitsSizeReader := fast.NewReader(bitsSizeBuf)
	bitsSize := readUint64Compact(bitsSizeReader)
	raw = raw[:len(raw)-bitsSizeReader.Position()]

	if uint64(len(raw)) < bitsSize {
		err = ErrMalformedEncoding
		return
	}

	bbits = &bits.Array{Bytes: raw[uint64(len(raw))-bitsSize:]}
	bbytes = raw[:uint64(len(raw))-bitsSize]
	return
}

func UnmarshalBinaryAdapter(raw []byte, unmarshalCser func(reader *Reader) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrMalformedEncoding
		}
	}()

	bbits, bbytes, err := binaryToCSER(raw)
	if err != nil {
		return err
	}

	bodyReader := &Reader{
		BitsR:  bits.NewReader(bbits),
		BytesR: fast.NewReader(bbytes),
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
