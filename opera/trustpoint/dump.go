package trustpoint

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/status-im/keycard-go/hexutils"

	"github.com/Fantom-foundation/go-opera/utils/iodb"
	"github.com/Fantom-foundation/go-opera/utils/ioread"
)

var (
	fileHeader  = hexutils.HexToBytes("642b00ac")
	fileVersion = hexutils.HexToBytes("00010001")
)

func checkFileHeader(reader io.Reader) error {
	headerAndVersion := make([]byte, len(fileHeader)+len(fileVersion))
	err := ioread.ReadAll(reader, headerAndVersion)
	if err != nil {
		return err
	}
	if bytes.Compare(headerAndVersion[:len(fileHeader)], fileHeader) != 0 {
		return fmt.Errorf("expected a genesis file, mismatched file header")
	}
	if bytes.Compare(headerAndVersion[len(fileHeader):], fileVersion) != 0 {
		got := hexutils.BytesToHex(headerAndVersion[len(fileHeader):])
		expected := hexutils.BytesToHex(fileVersion)
		return fmt.Errorf("wrong version of trustpoint file, got=%s, expected=%s", got, expected)
	}
	return nil
}

func OpenStore(input io.Reader) (h hash.Hash, read func(*Store) error, err error) {
	err = checkFileHeader(input)
	if err != nil {
		return
	}
	err = ioread.ReadAll(input, h[:])
	if err != nil {
		return
	}
	read = func(s *Store) error {
		gzipReader, err := gzip.NewReader(input)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		err = s.readFrom(gzipReader)
		if err != nil {
			return err
		}
		return nil
	}
	return
}

func (s *Store) readFrom(r io.Reader) error {
	return iodb.Read(r, s.db.NewBatch())
}

func WriteStore(output io.Writer, s *Store) error {
	_, err := output.Write(append(fileHeader, fileVersion...))
	if err != nil {
		return err
	}
	h := s.Hash()
	_, err = output.Write(h[:])
	if err != nil {
		return err
	}
	gzipWriter := gzip.NewWriter(output)
	defer gzipWriter.Close()
	err = s.writeTo(gzipWriter)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) writeTo(w io.Writer) error {
	return iodb.Write(w, s.db)
}
