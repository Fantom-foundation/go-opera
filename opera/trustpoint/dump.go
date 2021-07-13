package trustpoint

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

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

func ReadStore(input io.Reader, s *Store) error {
	err := checkFileHeader(input)
	if err != nil {
		return err
	}
	gzipReader, err := gzip.NewReader(input)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	_, err = gzipReader.Read(s.GenesisHash[:])
	if err != nil {
		return err
	}

	return iodb.Read(gzipReader, s.db.NewBatch())
}

func WriteStore(s *Store, output io.Writer) error {
	_, err := output.Write(append(fileHeader, fileVersion...))
	if err != nil {
		return err
	}

	gzipWriter := gzip.NewWriter(output)
	defer gzipWriter.Close()

	_, err = gzipWriter.Write(s.GenesisHash[:])
	if err != nil {
		return err
	}

	return iodb.Write(gzipWriter, s.db)
}
